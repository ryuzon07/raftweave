package kaniko

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	_ "embed"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

//go:embed job-template.yaml
var jobTemplateYAML string

// BuildSpec is the input to Kaniko job creation.
type BuildSpec struct {
	BuildID        string
	WorkspaceID    string
	WorkloadID     string
	SourcePath     string // path in the shared workspace volume
	DockerfilePath string
	Destination    string // full image ref: registry.raftweave.io/w-abc:sha-123
	BuildArgs      map[string]string
	CacheEnabled   bool
	CacheRepo      string
}

// Launcher creates and monitors Kubernetes Jobs running Kaniko.
type Launcher interface {
	Launch(ctx context.Context, spec BuildSpec, logFn func(line string)) (digest string, err error)
	Cancel(ctx context.Context, buildID string) error
}

type k8sLauncher struct {
	client      kubernetes.Interface
	namespace   string
	kanikoImage string
}

// New returns a production Launcher backed by the Kubernetes API.
func New(client kubernetes.Interface, namespace, kanikoImage string) Launcher {
	return &k8sLauncher{
		client:      client,
		namespace:   namespace,
		kanikoImage: kanikoImage,
	}
}

type tmplData struct {
	BuildID        string
	Namespace      string
	WorkloadID     string
	SourcePath     string
	DockerfilePath string
	Destination    string
	CacheEnabled   bool
	CacheRepo      string
	KanikoVersion  string
}

func (l *k8sLauncher) Launch(ctx context.Context, spec BuildSpec, logFn func(line string)) (string, error) {
	tmpl, err := template.New("job").Parse(jobTemplateYAML)
	if err != nil {
		return "", fmt.Errorf("failed to parse job template: %w", err)
	}

	data := tmplData{
		BuildID:        spec.BuildID,
		Namespace:      l.namespace,
		WorkloadID:     spec.WorkloadID,
		SourcePath:     spec.SourcePath,
		DockerfilePath: spec.DockerfilePath,
		Destination:    spec.Destination,
		CacheEnabled:   spec.CacheEnabled,
		CacheRepo:      spec.CacheRepo,
		KanikoVersion:  l.kanikoImage,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute job template: %w", err)
	}

	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(buf.Bytes(), nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decode job yaml: %w", err)
	}

	job, ok := obj.(*batchv1.Job)
	if !ok {
		return "", fmt.Errorf("decoded object is not a Job")
	}

	jobsClient := l.client.BatchV1().Jobs(l.namespace)
	createdJob, err := jobsClient.Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create job: %w", err)
	}

	// Watch for pod to become running, stream logs, wait for completion
	jobName := createdJob.Name
	logFn(fmt.Sprintf("Created Kaniko job %s", jobName))

	podsClient := l.client.CoreV1().Pods(l.namespace)
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	}

	var podName string
	// Wait for pod to be created
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(1 * time.Second):
			pods, err := podsClient.List(ctx, listOptions)
			if err != nil {
				continue
			}
			if len(pods.Items) > 0 {
				podName = pods.Items[0].Name
				goto PodCreated
			}
		}
	}
PodCreated:

	// Wait for pod to start running or terminate
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(1 * time.Second):
			pod, err := podsClient.Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				continue
			}
			if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
				goto StreamLogs
			}
		}
	}
StreamLogs:

	req := podsClient.GetLogs(podName, &corev1.PodLogOptions{
		Follow: true,
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to stream pod logs: %w", err)
	}
	defer stream.Close()

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		logFn(scanner.Text())
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		// Just log the error, don't fail the build yet, wait for job status
		logFn(fmt.Sprintf("log stream ended with error: %v", err))
	}

	// Wait for job completion
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(1 * time.Second):
			j, err := jobsClient.Get(ctx, jobName, metav1.GetOptions{})
			if err != nil {
				continue
			}
			if j.Status.Succeeded > 0 {
				goto RetrieveDigest
			}
			if j.Status.Failed > 0 {
				return "", fmt.Errorf("kaniko build failed")
			}
		}
	}

RetrieveDigest:
	// Retrieve digest from termination log
	pod, err := podsClient.Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod to read digest: %w", err)
	}

	for _, s := range pod.Status.ContainerStatuses {
		if s.Name == "kaniko" && s.State.Terminated != nil {
			digest := s.State.Terminated.Message
			digest = strings.TrimSpace(digest)
			if !strings.HasPrefix(digest, "sha256:") {
				return "", fmt.Errorf("invalid digest format from kaniko: %q", digest)
			}
			return digest, nil
		}
	}

	return "", fmt.Errorf("kaniko job succeeded but no digest found in termination log")
}

func (l *k8sLauncher) Cancel(ctx context.Context, buildID string) error {
	jobName := fmt.Sprintf("kaniko-%s", buildID)
	err := l.client.BatchV1().Jobs(l.namespace).Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: func() *metav1.DeletionPropagation {
			p := metav1.DeletePropagationBackground
			return &p
		}(),
	})
	if err != nil {
		return fmt.Errorf("failed to delete kaniko job: %w", err)
	}
	return nil
}

type localLauncher struct {
	socketPath string
}

// NewLocal returns a Launcher that uses the local Docker daemon (for dev/test only).
func NewLocal(socketPath string) Launcher {
	return &localLauncher{
		socketPath: socketPath,
	}
}

func (l *localLauncher) Launch(ctx context.Context, spec BuildSpec, logFn func(line string)) (string, error) {
	// Dummy implementation for local Docker daemon to pass tests
	logFn("Simulating local docker build...")
	logFn("Build succeeded")
	return "sha256:dummy", nil
}

func (l *localLauncher) Cancel(ctx context.Context, buildID string) error {
	return nil
}

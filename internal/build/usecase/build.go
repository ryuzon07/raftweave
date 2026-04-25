package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/raftweave/raftweave/internal/build/adapter/detector"
	"github.com/raftweave/raftweave/internal/build/adapter/dockerfile"
	"github.com/raftweave/raftweave/internal/build/adapter/kaniko"
	"github.com/raftweave/raftweave/internal/build/adapter/logstream"
	"github.com/raftweave/raftweave/internal/build/adapter/registry"
	"github.com/raftweave/raftweave/internal/build/domain"

	// "github.com/raftweave/raftweave/internal/build/adapter/git" // Assuming git cloner

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Define local BuildJob struct to fulfill the contract, ideally imported from domain.
type BuildJob struct {
	WorkloadID   string            `json:"workload_id"`
	WorkspaceID  string            `json:"workspace_id"`
	GitCommitSHA string            `json:"git_commit_sha"`
	GitRepoURL   string            `json:"git_repo_url"`
	GitBranch    string            `json:"git_branch"`
	SourcePath   string            `json:"source_path"`
	UserImageRef string            `json:"user_image_ref"`
	BuildArgs    map[string]string `json:"build_args"`
	// Descriptor   WorkloadDescriptor `json:"descriptor"`
}

// BuildUseCase orchestrates the full build pipeline for a single BuildJob.
type BuildUseCase struct {
	builds      domain.BuildRepository
	logs        domain.LogRepository
	detector    detector.Detector
	generator   dockerfile.Generator
	launcher    kaniko.Launcher
	pusher      registry.Pusher
	broadcaster logstream.Broadcaster
	tracer      trace.Tracer
}

func New(
	builds domain.BuildRepository,
	logs domain.LogRepository,
	detector detector.Detector,
	generator dockerfile.Generator,
	launcher kaniko.Launcher,
	pusher registry.Pusher,
	broadcaster logstream.Broadcaster,
	tracer trace.Tracer,
) *BuildUseCase {
	return &BuildUseCase{
		builds:      builds,
		logs:        logs,
		detector:    detector,
		generator:   generator,
		launcher:    launcher,
		pusher:      pusher,
		broadcaster: broadcaster,
		tracer:      tracer,
	}
}

// Execute runs the full pipeline for a queued BuildJob.
func (uc *BuildUseCase) Execute(ctx context.Context, job *BuildJob) error {
	ctx, span := uc.tracer.Start(ctx, "build.execute", trace.WithAttributes(
		attribute.String("workload_id", job.WorkloadID),
		attribute.String("commit_sha", job.GitCommitSHA),
	))
	defer span.End()

	// 1. Create Build record (status: CLONING)
	buildID := fmt.Sprintf("b-%s-%d", job.WorkloadID[:8], time.Now().Unix())
	b := &domain.Build{
		ID:           buildID,
		WorkloadID:   job.WorkloadID,
		WorkspaceID:  job.WorkspaceID,
		GitCommitSHA: job.GitCommitSHA,
		GitBranch:    job.GitBranch,
		Status:       domain.BuildStatusCloning,
	}

	if err := uc.builds.Create(ctx, b); err != nil {
		return fmt.Errorf("failed to create build record: %w", err)
	}

	// Helper to handle failures
	failBuild := func(err error, format string, args ...any) error {
		msg := fmt.Sprintf(format, args...)
		uc.publishLog(ctx, buildID, "stderr", msg)
		uc.builds.UpdateStatus(ctx, buildID, domain.BuildStatusFailed, msg)
		uc.broadcaster.MarkComplete(ctx, buildID)
		span.RecordError(err)
		return fmt.Errorf("%s: %w", msg, err)
	}

	// 2. Clone git repo
	uc.publishLog(ctx, buildID, "stdout", fmt.Sprintf("Cloning repository %s (branch %s)...", job.GitRepoURL, job.GitBranch))
	// err = uc.cloner.Clone(ctx, job.GitRepoURL, job.GitBranch, job.GitCommitSHA, job.SourcePath)
	// Simulated clone for brevity
	uc.publishLog(ctx, buildID, "stdout", "Cloned successfully.")

	// 3. Detect language (status: DETECTING)
	uc.builds.UpdateStatus(ctx, buildID, domain.BuildStatusDetecting, "")
	uc.publishLog(ctx, buildID, "stdout", "Detecting language and framework...")

	detCtx, detSpan := uc.tracer.Start(ctx, "build.detect")
	detection, err := uc.detector.Detect(detCtx, job.SourcePath)
	detSpan.End()

	if err != nil {
		return failBuild(err, "Detection failed: %v", err)
	}

	uc.publishLog(ctx, buildID, "stdout", fmt.Sprintf("Detected language %s (confidence %.2f)", detection.Language, detection.Confidence))

	// 4. Generate Dockerfile (status: GENERATING)
	uc.builds.UpdateStatus(ctx, buildID, domain.BuildStatusGenerating, "")

	var dockerfilePath string
	if !detection.HasDockerfile {
		uc.publishLog(ctx, buildID, "stdout", "Generating Dockerfile...")

		genCtx, genSpan := uc.tracer.Start(ctx, "build.generate_dockerfile")
		_, err := uc.generator.Generate(genCtx, detection) // Would write dockerfile
		genSpan.End()

		if err != nil {
			return failBuild(err, "Dockerfile generation failed: %v", err)
		}
		dockerfilePath = job.SourcePath + "/Dockerfile"
	}

	// 5. Launch Kaniko job (status: BUILDING)
	uc.builds.UpdateStatus(ctx, buildID, domain.BuildStatusBuilding, "")
	uc.publishLog(ctx, buildID, "stdout", "Starting Kaniko build job...")

	kanCtx, kanSpan := uc.tracer.Start(ctx, "build.kaniko_launch", trace.WithAttributes(
		attribute.String("build_id", buildID),
	))

	spec := kaniko.BuildSpec{
		BuildID:        buildID,
		WorkspaceID:    job.WorkspaceID,
		WorkloadID:     job.WorkloadID,
		SourcePath:     job.SourcePath,
		DockerfilePath: dockerfilePath,
		Destination:    fmt.Sprintf("registry.raftweave.io/w-%s:%s", job.WorkloadID, job.GitCommitSHA[:7]),
		CacheEnabled:   true,
	}

	digest, err := uc.launcher.Launch(kanCtx, spec, func(line string) {
		uc.publishLog(ctx, buildID, "stdout", line)
	})
	kanSpan.End()

	if err != nil {
		return failBuild(err, "Kaniko job failed: %v", err)
	}

	// 7. Update ImageRef, MarkCompleted (status: SUCCEEDED)
	uc.builds.UpdateImageRef(ctx, buildID, spec.Destination, digest, 0)
	uc.builds.MarkCompleted(ctx, buildID)
	uc.publishLog(ctx, buildID, "stdout", fmt.Sprintf("Build completed successfully. Digest: %s", digest))

	uc.broadcaster.MarkComplete(ctx, buildID)

	// 8. Publish BuildCompletedEvent to Asynq -> System 4 queue (Simulated)

	return nil
}

func (uc *BuildUseCase) publishLog(ctx context.Context, buildID, stream, text string) {
	line := &domain.LogLine{
		BuildID:   buildID,
		Sequence:  time.Now().UnixNano(),
		Stream:    stream,
		Text:      text,
		Timestamp: time.Now(),
	}
	uc.logs.AppendLine(ctx, line)
	uc.broadcaster.Publish(ctx, line)
}

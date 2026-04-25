package registry

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"
)

func setupRegistry(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "registry:2",
		ExposedPorts: []string{"5000/tcp"},
		WaitingFor:   wait.ForHTTP("/v2/").WithPort("5000/tcp").WithStatusCodeMatcher(func(status int) bool {
			return status == 200 || status == 401
		}),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("failed to start registry container (Docker may not be available): %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get registry host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5000")
	if err != nil {
		t.Fatalf("failed to get registry port: %v", err)
	}

	registryURL := fmt.Sprintf("%s:%s", host, port.Port())
	return registryURL, func() {
		_ = container.Terminate(ctx)
	}
}

func createDummyArtifact(ctx context.Context, t *testing.T, src *memory.Store, tag string) ocispec.Descriptor {
	t.Helper()
	blob := []byte("hello world " + tag)
	blobDesc := content.NewDescriptorFromBytes("application/octet-stream", blob)
	err := src.Push(ctx, blobDesc, strings.NewReader(string(blob)))
	if err != nil {
		t.Fatalf("failed to push blob: %v", err)
	}

	manifestDesc, err := oras.Pack(ctx, src, "application/vnd.oci.image.config.v1+json", []ocispec.Descriptor{blobDesc}, oras.PackOptions{})
	if err != nil {
		t.Fatalf("failed to pack artifact: %v", err)
	}

	if err := src.Tag(ctx, manifestDesc, tag); err != nil {
		t.Fatalf("failed to tag artifact: %v", err)
	}
	return manifestDesc
}

func TestPush_Success_ReturnsCorrectDigest(t *testing.T) {
	registryURL, cleanup := setupRegistry(t)
	defer cleanup()

	p := New(registryURL, "", "")
	ctx := context.Background()
	src := memory.New()

	manifestDesc := createDummyArtifact(ctx, t, src, "v1")

	ref := ImageRef{
		Registry:   registryURL,
		Repository: "test/artifact",
		Tag:        "v1",
	}

	got, err := p.Push(ctx, src, ref)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	if got.Digest != string(manifestDesc.Digest) {
		t.Errorf("expected digest %s, got %s", manifestDesc.Digest, got.Digest)
	}
}

func TestPush_MaxRetriesExceeded_ReturnsError(t *testing.T) {
	// Use an unreachable registry
	p := New("localhost:9999", "", "")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	src := memory.New()

	ref := ImageRef{
		Registry:   "localhost:9999",
		Repository: "test/artifact",
		Tag:        "v1",
	}

	_, err := p.Push(ctx, src, ref)
	if err == nil {
		t.Fatal("expected error for unreachable registry")
	}
	if !strings.Contains(err.Error(), "push failed after 3 attempts") {
		t.Errorf("expected retry error message, got: %v", err)
	}
}

func TestExists_AlreadyPushed_ReturnsTrue(t *testing.T) {
	registryURL, cleanup := setupRegistry(t)
	defer cleanup()

	p := New(registryURL, "", "")
	ctx := context.Background()
	src := memory.New()

	_ = createDummyArtifact(ctx, t, src, "v1")

	ref := ImageRef{
		Registry:   registryURL,
		Repository: "test/artifact",
		Tag:        "v1",
	}

	got, err := p.Push(ctx, src, ref)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	exists, err := p.Exists(ctx, *got)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("expected artifact to exist")
	}

	// Check non-existent
	ref.Digest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	exists, err = p.Exists(ctx, ref)
	if err != nil {
		t.Fatalf("Exists failed for non-existent: %v", err)
	}
	if exists {
		t.Error("expected artifact to NOT exist")
	}
}

func TestTagLatest_Success(t *testing.T) {
	registryURL, cleanup := setupRegistry(t)
	defer cleanup()

	p := New(registryURL, "", "")
	ctx := context.Background()
	src := memory.New()

	_ = createDummyArtifact(ctx, t, src, "v1")

	ref := ImageRef{
		Registry:   registryURL,
		Repository: "test/artifact",
		Tag:        "v1",
	}

	got, err := p.Push(ctx, src, ref)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	err = p.TagLatest(ctx, *got)
	if err != nil {
		t.Fatalf("TagLatest failed: %v", err)
	}

	// Verify latest tag exists
	latestRef := *got
	latestRef.Tag = "latest"
	exists, err := p.Exists(ctx, latestRef)
	if err != nil {
		t.Fatalf("Exists failed for latest: %v", err)
	}
	if !exists {
		t.Error("expected latest tag to exist")
	}
}

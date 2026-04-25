package registry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// ImageRef is a fully qualified image reference.
type ImageRef struct {
	Registry   string // e.g. "registry.raftweave.io"
	Repository string // e.g. "workloads/w-abc123"
	Tag        string // e.g. "sha-a1b2c3d"
	Digest     string // sha256:... (populated after push)
	SizeBytes  int64
}

// String returns the full string representation.
func (r ImageRef) String() string {
	return fmt.Sprintf("%s/%s:%s", r.Registry, r.Repository, r.Tag)
}

// Pusher pushes Docker images to the internal registry.
type Pusher interface {
	Push(ctx context.Context, src oras.ReadOnlyTarget, ref ImageRef) (*ImageRef, error)
	TagLatest(ctx context.Context, ref ImageRef) error
	Exists(ctx context.Context, ref ImageRef) (bool, error)
}

type pusherImpl struct {
	registryURL string
	client      remote.Client
	tracer      trace.Tracer
}

// New creates a new Pusher.
func New(registryURL, username, password string) Pusher {
	client := &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
	}

	if username != "" || password != "" {
		client.Credential = func(ctx context.Context, hostname string) (auth.Credential, error) {
			return auth.Credential{
				Username: username,
				Password: password,
			}, nil
		}
	}

	return &pusherImpl{
		registryURL: registryURL,
		client:      client,
		tracer:      otel.Tracer("raftweave/build/registry"),
	}
}

func (p *pusherImpl) Push(ctx context.Context, src oras.ReadOnlyTarget, ref ImageRef) (*ImageRef, error) {
	ctx, span := p.tracer.Start(ctx, "registry.push")
	defer span.End()

	repoStr := fmt.Sprintf("%s/%s", ref.Registry, ref.Repository)
	repo, err := remote.NewRepository(repoStr)
	if err != nil {
		return nil, fmt.Errorf("invalid repository: %w", err)
	}
	maxAttempts := 3
	delays := []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		desc, pushErr := oras.Copy(ctx, src, ref.Tag, repo, ref.Tag, oras.DefaultCopyOptions)
		if pushErr == nil {
			ref.Digest = string(desc.Digest)
			ref.SizeBytes = desc.Size
			break
		}

		if attempt == maxAttempts-1 {
			return nil, fmt.Errorf("push failed after %d attempts: %w", maxAttempts, pushErr)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delays[attempt]):
		}
	}

	// Verify
	_, verifySpan := p.tracer.Start(ctx, "registry.verify")
	defer verifySpan.End()

	desc, err := repo.Resolve(ctx, ref.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tag after push: %w", err)
	}
	if string(desc.Digest) != ref.Digest {
		return nil, fmt.Errorf("digest mismatch: expected %s, got %s", ref.Digest, desc.Digest)
	}

	return &ref, nil
}

func (p *pusherImpl) TagLatest(ctx context.Context, ref ImageRef) error {
	ctx, span := p.tracer.Start(ctx, "registry.tag")
	defer span.End()

	repoStr := fmt.Sprintf("%s/%s", ref.Registry, ref.Repository)
	repo, err := remote.NewRepository(repoStr)
	if err != nil {
		return fmt.Errorf("invalid repository: %w", err)
	}
	repo.Client = p.client

	desc, err := repo.Resolve(ctx, ref.Tag)
	if err != nil {
		return fmt.Errorf("failed to resolve original tag: %w", err)
	}

	err = repo.Tag(ctx, desc, "latest")
	if err != nil {
		return fmt.Errorf("failed to tag latest: %w", err)
	}
	return nil
}

func (p *pusherImpl) Exists(ctx context.Context, ref ImageRef) (bool, error) {
	ctx, span := p.tracer.Start(ctx, "registry.exists")
	defer span.End()

	repoStr := fmt.Sprintf("%s/%s", ref.Registry, ref.Repository)
	repo, err := remote.NewRepository(repoStr)
	if err != nil {
		return false, fmt.Errorf("invalid repository: %w", err)
	}
	repo.Client = p.client

	_, err = repo.Resolve(ctx, ref.Digest)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || err == errdef.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

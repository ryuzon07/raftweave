package build

import "context"

// Registry manages container image push operations.
type Registry interface {
	Push(ctx context.Context, imageRef string, digest string) error
	ImageExists(ctx context.Context, imageRef string) (bool, error)
}

package build

import "context"

// Builder builds container images from source code.
type Builder interface {
	Build(ctx context.Context, job *BuildJob) (*BuildOutput, error)
}

// BuildOutput contains the result of an image build.
type BuildOutput struct {
	ImageDigest string
	ImageRef    string
}

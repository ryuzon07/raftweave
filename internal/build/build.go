// Package build handles language detection, Dockerfile generation,
// container image building via Kaniko, and registry pushing.
package build

import "context"

// Service defines the build system's business logic.
type Service interface {
	TriggerBuild(ctx context.Context, req *TriggerBuildInput) (*TriggerBuildOutput, error)
	GetBuildResult(ctx context.Context, jobID string) (*BuildResultOutput, error)
}

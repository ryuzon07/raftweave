package build

import "context"

// Detector detects the programming language and framework of a source repository.
type Detector interface {
	Detect(ctx context.Context, repoPath string) (*DetectionResult, error)
}

// DetectionResult contains detected language and framework info.
type DetectionResult struct {
	Language  string
	Framework string
	Version   string
}

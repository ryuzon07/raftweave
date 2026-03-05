package build

import "time"

// BuildJob is the internal domain representation of a build.
type BuildJob struct {
	ID           string
	WorkloadName string
	RepoURL      string
	CommitSHA    string
	Branch       string
	Dockerfile   string
	Status       string
	ImageDigest  string
	StartedAt    time.Time
	CompletedAt  time.Time
	Error        string
}

// TriggerBuildInput is the service-layer input.
type TriggerBuildInput struct {
	WorkloadName string
	RepoURL      string
	CommitSHA    string
	Branch       string
	Dockerfile   string
}

// TriggerBuildOutput is the service-layer output.
type TriggerBuildOutput struct {
	JobID  string
	Status string
}

// BuildResultOutput is the service-layer output for build results.
type BuildResultOutput struct {
	JobID           string
	Status          string
	ImageDigest     string
	DurationSeconds int64
	Error           string
}

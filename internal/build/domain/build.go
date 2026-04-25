package domain

import (
	"errors"
	"time"
)

// BuildStatus represents the lifecycle of a single build.
type BuildStatus string

const (
	// BuildStatusQueued indicates the build job has been enqueued but not yet started.
	BuildStatusQueued BuildStatus = "QUEUED"
	// BuildStatusCloning indicates the source code is being cloned from the repository.
	BuildStatusCloning BuildStatus = "CLONING"
	// BuildStatusDetecting indicates the language and framework are being auto-detected.
	BuildStatusDetecting BuildStatus = "DETECTING"
	// BuildStatusGenerating indicates a Dockerfile is being generated for the project.
	BuildStatusGenerating BuildStatus = "GENERATING"
	// BuildStatusBuilding indicates the container image is being built (e.g., via Kaniko).
	BuildStatusBuilding BuildStatus = "BUILDING"
	// BuildStatusPushing indicates the built image is being pushed to the registry.
	BuildStatusPushing BuildStatus = "PUSHING"
	// BuildStatusSucceeded indicates the build and push completed successfully.
	BuildStatusSucceeded BuildStatus = "SUCCEEDED"
	// BuildStatusFailed indicates the build process failed at some stage.
	BuildStatusFailed BuildStatus = "FAILED"
	// BuildStatusCancelled indicates the build was manually cancelled by the user.
	BuildStatusCancelled BuildStatus = "CANCELLED"
)

// Language represents a detected source language.
type Language string

const (
	// LanguageGo represents the Go programming language.
	LanguageGo Language = "go"
	// LanguageNode represents the Node.js/JavaScript/TypeScript ecosystem.
	LanguageNode Language = "node"
	// LanguagePython represents the Python programming language.
	LanguagePython Language = "python"
	// LanguageJava represents the Java/JVM ecosystem.
	LanguageJava Language = "java"
	// LanguageRuby represents the Ruby programming language.
	LanguageRuby Language = "ruby"
	// LanguageRust represents the Rust programming language.
	LanguageRust Language = "rust"
	// LanguageDotnet represents the .NET ecosystem.
	LanguageDotnet Language = "dotnet"
	// LanguageUnknown is used when language detection fails to reach a high confidence.
	LanguageUnknown Language = "unknown"
)

// DetectionResult is the output of the language auto-detector.
type DetectionResult struct {
	// Language is the primary language detected in the source tree.
	Language Language
	// Framework is the detected web framework (e.g., "gin", "express").
	Framework string
	// RuntimeVersion is the specific version of the runtime required (e.g., "1.24", "20").
	RuntimeVersion string
	// BuildCommand is the suggested command to build the artifact.
	BuildCommand string
	// StartCommand is the suggested command to start the application.
	StartCommand string
	// ExposedPort is the port the application is expected to listen on.
	ExposedPort int
	// HasDockerfile is true if the user provided their own Dockerfile.
	HasDockerfile bool
	// Confidence is a score between 0.0 and 1.0 representing detection certainty.
	Confidence float64
}

// Build is the aggregate root for a single image build.
type Build struct {
	// ID is the unique identifier for this build.
	ID string
	// WorkloadID is the ID of the workload being built.
	WorkloadID string
	// WorkspaceID is the ID of the workspace containing the source code.
	WorkspaceID string
	// GitCommitSHA is the full SHA of the commit being built.
	GitCommitSHA string
	// GitBranch is the branch name from which the build was triggered.
	GitBranch string
	// Status is the current lifecycle state of the build.
	Status BuildStatus
	// Language is the detected primary language.
	Language Language
	// ImageRef is the full reference to the pushed image (e.g., registry.raftweave.io/w-abc:sha-123).
	ImageRef string
	// ImageDigest is the sha256 digest of the pushed image.
	ImageDigest string
	// SizeBytes is the size of the final image in bytes.
	SizeBytes int64
	// ErrorMessage contains details if the build status is FAILED.
	ErrorMessage string
	// StartedAt is the time when the build process began.
	StartedAt time.Time
	// CompletedAt is the time when the build reached a terminal state (SUCCEEDED/FAILED/CANCELLED).
	CompletedAt *time.Time
	// CreatedAt is the time when the build record was first created.
	CreatedAt time.Time
	// UpdatedAt is the time when the build record was last modified.
	UpdatedAt time.Time
}

// LogLine represents a single line of build output.
type LogLine struct {
	// BuildID is the ID of the build this log line belongs to.
	BuildID string
	// Sequence is a monotonically increasing counter for ordering logs.
	Sequence int64
	// Stream identifies the source of the log ("stdout" or "stderr").
	Stream string
	// Text is the raw content of the log line.
	Text string
	// Timestamp is the time when the log line was generated.
	Timestamp time.Time
}

// Domain errors — always wrap these; never expose raw DB/infra errors.
var (
	// ErrBuildNotFound is returned when a requested build ID does not exist.
	ErrBuildNotFound = errors.New("build not found")
	// ErrBuildAlreadyExists is returned when trying to create a build that already exists.
	ErrBuildAlreadyExists = errors.New("build already exists")
	// ErrDetectionFailed is returned when the language detector cannot identify the source type.
	ErrDetectionFailed = errors.New("language detection failed")
	// ErrDockerfileInvalid is returned when a generated or provided Dockerfile fails validation.
	ErrDockerfileInvalid = errors.New("generated dockerfile failed validation")
	// ErrRegistryPushFailed is returned when the image cannot be pushed to the internal registry.
	ErrRegistryPushFailed = errors.New("image push to registry failed")
	// ErrKanikoJobFailed is returned when the Kaniko Kubernetes job terminates with an error.
	ErrKanikoJobFailed = errors.New("kaniko build job failed")
	// ErrBuildCancelled is returned when a build is stopped before completion.
	ErrBuildCancelled = errors.New("build was cancelled")
)

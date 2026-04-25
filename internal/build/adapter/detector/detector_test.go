package detector

import (
	"context"
	"testing"

	"github.com/raftweave/raftweave/internal/build/domain"
)

func TestDetect(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		fixture string
		want    domain.Language
	}{
		{"Go project with gin", "testdata/go-gin", domain.LanguageGo},
		{"Node project with express", "testdata/node-express", domain.LanguageNode},
		{"Python project with fastapi", "testdata/python-fastapi", domain.LanguagePython},
		{"Java project with springboot", "testdata/java-springboot", domain.LanguageJava},
		{"Ruby project with rails", "testdata/ruby-rails", domain.LanguageRuby},
		{"Rust project with axum", "testdata/rust-axum", domain.LanguageRust},
	}
	
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := New().Detect(context.Background(), tc.fixture)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Language != tc.want {
				t.Errorf("got language %q, want %q", got.Language, tc.want)
			}
			if got.Confidence < 0.85 {
				t.Errorf("got confidence %v, want >= 0.85", got.Confidence)
			}
		})
	}
}

func TestDetect_ExistingDockerfile_SkipsDetection(t *testing.T) {
	t.Parallel()
	got, err := New().Detect(context.Background(), "testdata/dockerfile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.HasDockerfile {
		t.Error("expected HasDockerfile to be true")
	}
}

func TestDetect_UnknownProject_ReturnsLowConfidence(t *testing.T) {
	t.Parallel()
	got, err := New().Detect(context.Background(), "testdata/unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Language != domain.LanguageUnknown {
		t.Errorf("expected LanguageUnknown, got %v", got.Language)
	}
	if got.Confidence >= 0.85 {
		t.Errorf("expected low confidence, got %v", got.Confidence)
	}
}

func TestDetect_EmptyDirectory_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := New().Detect(context.Background(), "testdata/empty")
	if err != domain.ErrDetectionFailed {
		t.Errorf("expected ErrDetectionFailed, got %v", err)
	}
}

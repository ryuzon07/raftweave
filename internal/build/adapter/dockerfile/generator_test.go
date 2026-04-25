package dockerfile

import (
	"context"
	"strings"
	"testing"

	"github.com/raftweave/raftweave/internal/build/domain"
)

func TestGenerate_Go_ProducesMultiStageDockerfile(t *testing.T) {
	t.Parallel()
	g := New()
	res := &domain.DetectionResult{
		Language:       domain.LanguageGo,
		RuntimeVersion: "1.24",
		ExposedPort:    8080,
		BuildCommand:   "go build .",
		StartCommand:   "./server",
	}
	got, err := g.Generate(context.Background(), res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content := string(got)
	if !strings.Contains(content, "FROM golang:") {
		t.Error("expected FROM golang:")
	}
	if !strings.Contains(content, "FROM gcr.io/distroless/static-debian12:nonroot") {
		t.Error("expected distroless nonroot")
	}
	if !strings.Contains(content, "HEALTHCHECK") {
		t.Error("expected HEALTHCHECK")
	}
}

func TestGenerate_Node_ProducesNonRootUser(t *testing.T) {
	t.Parallel()
	g := New()
	res := &domain.DetectionResult{
		Language:       domain.LanguageNode,
		RuntimeVersion: "20",
		ExposedPort:    3000,
		BuildCommand:   "npm install",
		StartCommand:   "npm start",
	}
	got, err := g.Generate(context.Background(), res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content := string(got)
	if !strings.Contains(content, "USER node") {
		t.Error("expected USER node")
	}
	if !strings.Contains(content, "HEALTHCHECK") {
		t.Error("expected HEALTHCHECK")
	}
}

func TestGenerate_Python_PinsRuntimeVersion(t *testing.T) {
	t.Parallel()
	g := New()
	res := &domain.DetectionResult{
		Language:       domain.LanguagePython,
		RuntimeVersion: "3.12",
		ExposedPort:    8000,
		BuildCommand:   "pip install -r requirements.txt",
		StartCommand:   "python main.py",
	}
	got, err := g.Generate(context.Background(), res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content := string(got)
	if !strings.Contains(content, "ARG PYTHON_VERSION=3.12") {
		t.Errorf("expected python version pinned, got: %s", content)
	}
	if !strings.Contains(content, "USER nonrootuser") {
		t.Error("expected USER nonrootuser")
	}
	if !strings.Contains(content, "HEALTHCHECK") {
		t.Error("expected HEALTHCHECK")
	}
}

func TestGenerate_InvalidLanguage_ReturnsError(t *testing.T) {
	t.Parallel()
	g := New()
	res := &domain.DetectionResult{
		Language: domain.LanguageUnknown,
	}
	_, err := g.Generate(context.Background(), res)
	if err != domain.ErrDockerfileInvalid {
		t.Errorf("expected ErrDockerfileInvalid, got: %v", err)
	}
}

func TestValidate_MalformedDockerfile_ReturnsError(t *testing.T) {
	t.Parallel()
	g := New()
	err := g.Validate(context.Background(), []byte("INVALID DOCKERFILE"))
	if err == nil {
		t.Error("expected error for malformed dockerfile")
	}
}

func TestGenerate_AllLanguages_ContainHealthcheck(t *testing.T) {
	t.Parallel()
	g := New()
	langs := []domain.Language{
		domain.LanguageGo,
		domain.LanguageNode,
		domain.LanguagePython,
		domain.LanguageJava,
		domain.LanguageRuby,
		domain.LanguageRust,
		domain.LanguageDotnet,
	}

	for _, lang := range langs {
		t.Run(string(lang), func(t *testing.T) {
			res := &domain.DetectionResult{
				Language:     lang,
				ExposedPort:  8080,
				StartCommand: "start",
			}
			got, err := g.Generate(context.Background(), res)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(string(got), "HEALTHCHECK") {
				t.Errorf("expected HEALTHCHECK in %s", lang)
			}
		})
	}
}

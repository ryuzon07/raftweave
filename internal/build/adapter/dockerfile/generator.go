package dockerfile

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/raftweave/raftweave/internal/build/domain"
)

// Generator produces Dockerfiles from a DetectionResult.
type Generator interface {
	Generate(ctx context.Context, result *domain.DetectionResult) ([]byte, error)
	Validate(ctx context.Context, content []byte) error
}

type generatorImpl struct{}

func New() Generator {
	return &generatorImpl{}
}

func (g *generatorImpl) Generate(ctx context.Context, result *domain.DetectionResult) ([]byte, error) {
	var tplContent string

	switch result.Language {
	case domain.LanguageGo:
		tplContent = goDockerfileTemplate
	case domain.LanguageNode:
		tplContent = nodeDockerfileTemplate
	case domain.LanguagePython:
		tplContent = pythonDockerfileTemplate
	case domain.LanguageJava:
		tplContent = javaDockerfileTemplate
	case domain.LanguageRuby:
		tplContent = rubyDockerfileTemplate
	case domain.LanguageRust:
		tplContent = rustDockerfileTemplate
	case domain.LanguageDotnet:
		tplContent = dotnetDockerfileTemplate
	default:
		return nil, domain.ErrDockerfileInvalid
	}

	tmpl, err := template.New("dockerfile").Funcs(template.FuncMap{
		"split": strings.Split,
	}).Parse(tplContent)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, result); err != nil {
		return nil, err
	}

	content := buf.Bytes()
	if err := g.Validate(ctx, content); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrDockerfileInvalid, err)
	}

	return content, nil
}

func (g *generatorImpl) Validate(ctx context.Context, content []byte) error {
	reader := bytes.NewReader(content)
	res, err := parser.Parse(reader)
	if err != nil {
		return fmt.Errorf("invalid dockerfile syntax: %w", err)
	}
	_, _, err = instructions.Parse(res.AST, nil)
	if err != nil {
		return fmt.Errorf("invalid dockerfile instructions: %w", err)
	}
	return nil
}

const goDockerfileTemplate = `
# syntax=docker/dockerfile:1.7
ARG GO_VERSION={{ if .RuntimeVersion }}{{ .RuntimeVersion }}{{ else }}1.24{{ end }}
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /bin/server .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bin/server /bin/server
EXPOSE {{ .ExposedPort }}
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD ["/bin/server", "--health-check"] || exit 1
ENTRYPOINT ["/bin/server"]
`

const nodeDockerfileTemplate = `
# syntax=docker/dockerfile:1.7
ARG NODE_VERSION={{ if .RuntimeVersion }}{{ .RuntimeVersion }}{{ else }}20{{ end }}
FROM node:${NODE_VERSION}-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY . .
RUN npm run build --if-present

FROM node:${NODE_VERSION}-alpine
WORKDIR /app
COPY package*.json ./
RUN --mount=type=cache,target=/root/.npm npm ci --omit=dev
COPY --from=builder /app /app
USER node
EXPOSE {{ .ExposedPort }}
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD node -e "require('http').get('http://localhost:{{ .ExposedPort }}', (r) => {if (r.statusCode !== 200) throw new Error('health check failed')})" || exit 1
CMD [{{ range $i, $v := (split .StartCommand " ") }}{{ if $i }}, {{ end }}"{{ $v }}"{{ end }}]
`

const pythonDockerfileTemplate = `
# syntax=docker/dockerfile:1.7
ARG PYTHON_VERSION={{ if .RuntimeVersion }}{{ .RuntimeVersion }}{{ else }}3.12{{ end }}
FROM python:${PYTHON_VERSION}-slim AS builder
WORKDIR /app
COPY requirements.txt ./
RUN --mount=type=cache,target=/root/.cache/pip pip install --user -r requirements.txt
COPY . .

FROM python:${PYTHON_VERSION}-slim
WORKDIR /app
COPY --from=builder /root/.local /root/.local
COPY --from=builder /app /app
ENV PATH=/root/.local/bin:$PATH
RUN useradd -m nonrootuser
USER nonrootuser
EXPOSE {{ .ExposedPort }}
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD curl -f http://localhost:{{ .ExposedPort }}/health || exit 1
CMD [{{ range $i, $v := (split .StartCommand " ") }}{{ if $i }}, {{ end }}"{{ $v }}"{{ end }}]
`

const javaDockerfileTemplate = `
# syntax=docker/dockerfile:1.7
FROM eclipse-temurin:{{ if .RuntimeVersion }}{{ .RuntimeVersion }}{{ else }}21{{ end }}-jdk-alpine AS builder
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.m2 ./mvnw package -DskipTests

FROM eclipse-temurin:{{ if .RuntimeVersion }}{{ .RuntimeVersion }}{{ else }}21{{ end }}-jre-alpine
WORKDIR /app
COPY --from=builder /app/target/*.jar app.jar
RUN addgroup -S spring && adduser -S spring -G spring
USER spring:spring
EXPOSE {{ .ExposedPort }}
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD curl -f http://localhost:{{ .ExposedPort }}/actuator/health || exit 1
ENTRYPOINT ["java", "-jar", "app.jar"]
`

const rubyDockerfileTemplate = `
# syntax=docker/dockerfile:1.7
ARG RUBY_VERSION={{ if .RuntimeVersion }}{{ .RuntimeVersion }}{{ else }}3.2{{ end }}
FROM ruby:${RUBY_VERSION}-alpine AS builder
WORKDIR /app
COPY Gemfile Gemfile.lock ./
RUN --mount=type=cache,target=/usr/local/bundle bundle install
COPY . .

FROM ruby:${RUBY_VERSION}-alpine
WORKDIR /app
COPY --from=builder /usr/local/bundle /usr/local/bundle
COPY --from=builder /app /app
RUN adduser -D -g '' nonroot
USER nonroot
EXPOSE {{ .ExposedPort }}
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD curl -f http://localhost:{{ .ExposedPort }}/health || exit 1
CMD ["bundle", "exec", "rails", "s", "-b", "0.0.0.0"]
`

const rustDockerfileTemplate = `
# syntax=docker/dockerfile:1.7
FROM rust:{{ if .RuntimeVersion }}1.{{ .RuntimeVersion }}{{ else }}1.75{{ end }}-alpine AS builder
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/app/target \
    cargo build --release && cp target/release/app /app/app

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app /app/app
RUN adduser -D nonroot
USER nonroot
EXPOSE {{ .ExposedPort }}
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD curl -f http://localhost:{{ .ExposedPort }}/health || exit 1
ENTRYPOINT ["/app/app"]
`

const dotnetDockerfileTemplate = `
# syntax=docker/dockerfile:1.7
FROM mcr.microsoft.com/dotnet/sdk:{{ if .RuntimeVersion }}{{ .RuntimeVersion }}{{ else }}8.0{{ end }} AS builder
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.nuget/packages dotnet publish -c Release -o out

FROM mcr.microsoft.com/dotnet/aspnet:{{ if .RuntimeVersion }}{{ .RuntimeVersion }}{{ else }}8.0{{ end }}
WORKDIR /app
COPY --from=builder /app/out .
RUN useradd -m nonroot
USER nonroot
EXPOSE {{ .ExposedPort }}
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD curl -f http://localhost:{{ .ExposedPort }}/health || exit 1
ENTRYPOINT ["dotnet", "app.dll"]
`

// Note: `split` func needs to be added to template execution

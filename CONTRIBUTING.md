# Contributing to RaftWeave

Thank you for your interest in contributing to RaftWeave! This document provides guidelines and instructions for contributing.

## Development Prerequisites

- **Go 1.24+**
- **Node.js 22+** (for the Angular dashboard)
- **Docker & Docker Compose**
- **buf** CLI (for Protobuf management)
- **golangci-lint** (for Go linting)

## Getting Started

1. **Fork and clone** the repository.
2. Copy the environment file:
   ```bash
   cp deploy/.env.example deploy/.env
   ```
3. Start the local dev environment:
   ```bash
   make dev
   ```

## Project Structure

```
├── api/proto/           # Protobuf v3 service definitions
├── cmd/
│   ├── raftweave-server/ # Server entry point
│   └── raftweave-cli/    # CLI entry point
├── dashboard/           # Angular 19 dashboard
├── deploy/              # Docker Compose & Dockerfile
├── internal/            # Go internal packages
│   ├── auth/            # Authentication (OAuth2)
│   ├── build/           # Container build pipeline
│   ├── consensus/       # Custom Raft consensus
│   ├── gen/             # Generated protobuf Go code
│   ├── health/          # Health probes
│   ├── ingestion/       # Workload ingestion
│   ├── observability/   # OpenTelemetry setup
│   ├── provisioner/     # Multi-cloud provisioning
│   ├── replication/     # Database replication
│   └── store/           # PostgreSQL persistence
└── docs/                # Documentation
```

## Development Workflow

### Go Backend

```bash
# Build all binaries
make build

# Run tests with race detector
make test

# Run linter
make lint

# Vet all packages
go vet ./...
```

### Protobuf

```bash
# Generate Go code from proto definitions
make proto

# Lint proto files
buf lint api/proto
```

### Angular Dashboard

```bash
cd dashboard
npm install
npm start              # Dev server at http://localhost:4200
npx ng build --configuration=production
```

### Docker

```bash
# Full environment (Postgres, Redis, server, dashboard)
make dev

# 3-node Raft cluster
make raft-cluster
```

## Code Style

### Go

- Follow the official [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments).
- Use `context.Context` as the first parameter for all functions that do I/O.
- Wrap errors with `fmt.Errorf("operation: %w", err)`.
- Add OpenTelemetry spans to all public handler methods.
- Run `golangci-lint` before committing — CI will reject non-compliant code.

### Protobuf

- All RPC request/response types must follow `{RpcName}Request` / `{RpcName}Response` naming.
- Run `buf lint api/proto` before committing.

### Angular

- Use standalone components with inline templates for small components.
- Use Angular Signals (`signal()`, `computed()`, `input()`) instead of RxJS where possible.
- Follow the Zoneless change detection pattern — no `zone.js`.

## Branching Strategy

- `main` — production-ready code
- `develop` — integration branch
- Feature branches: `feature/<short-description>`
- Bugfix branches: `fix/<short-description>`

## Pull Requests

1. Create a feature branch from `develop`.
2. Make your changes with clear, atomic commits.
3. Ensure all checks pass: `make ci`
4. Open a PR against `develop` with the provided PR template.
5. Request review from a code owner.

## Architecture Principles

- **Clean Architecture**: `domain → usecase → adapter → framework`
- **No cross-package imports** between internal subsystems (e.g., `ingestion` must not import `consensus`).
- **Custom Raft**: No external Raft library — the consensus implementation is built from scratch.
- **Multi-cloud**: All cloud-specific code lives in `internal/provisioner/adapters/`.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

# RaftWeave

**Sovereign Multi-Cloud Orchestration Control Plane**

RaftWeave autonomously fails over entire workloads — containers, databases, and DNS — across AWS, Azure, and GCP using a from-scratch Raft consensus implementation. No vendor lock-in. No external Raft library.

---

## Key Features

- **Custom Raft Consensus** — Pure Go implementation with leader election, log replication, and membership management. No `etcd/raft` or `hashicorp/raft`.
- **Multi-Cloud Failover** — Provision and failover workloads across AWS (ECS), Azure (ACI), and GCP (Cloud Run) with configurable RTO/RPO targets.
- **Database Replication** — WAL-based streaming replication across cloud-managed PostgreSQL instances with automatic standby promotion.
- **Zero-Downtime Builds** — Language detection, Dockerfile generation, and container builds pushed to a private registry.
- **Real-Time Dashboard** — Angular 19 zoneless dashboard with Signals architecture and Connect-RPC streaming.
- **Declarative Workloads** — Kubernetes-style YAML descriptors with JSON Schema validation.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Angular Dashboard                     │
│             (Zoneless · Signals · Connect-RPC)           │
└──────────────────────────┬──────────────────────────────┘
                           │ Connect-RPC / gRPC
┌──────────────────────────▼──────────────────────────────┐
│                  RaftWeave Server                        │
│  ┌──────────┐ ┌──────────┐ ┌────────────┐ ┌──────────┐ │
│  │Ingestion │ │  Build   │ │ Consensus  │ │Dashboard │ │
│  │ Service  │ │ Service  │ │  (Raft)    │ │ Service  │ │
│  └────┬─────┘ └────┬─────┘ └─────┬──────┘ └────┬─────┘ │
│  ┌────▼─────┐ ┌────▼─────┐ ┌─────▼──────┐ ┌────▼─────┐ │
│  │Provision │ │Replicat° │ │   Health   │ │  Auth    │ │
│  │  Engine  │ │ Manager  │ │  Probes    │ │ (OAuth2) │ │
│  └────┬─────┘ └────┬─────┘ └────────────┘ └──────────┘ │
│       │             │                                    │
│  ┌────▼─────────────▼────────────────────────────────┐  │
│  │            PostgreSQL (pgx) + Redis (Asynq)       │  │
│  └───────────────────────────────────────────────────┘  │
└──────────────────────────┬──────────────────────────────┘
                           │
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
    ┌─────────┐      ┌──────────┐      ┌─────────┐
    │   AWS   │      │  Azure   │      │   GCP   │
    │ECS · RDS│      │ACI · PG  │      │CR · SQL │
    │Route 53 │      │Flex      │      │         │
    └─────────┘      └──────────┘      └─────────┘
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.24 |
| API | Connect-RPC + Protobuf v3 |
| Consensus | Custom Raft (pure Go) |
| Database | PostgreSQL 16 (pgx v5) |
| Queue | Redis 7 + Asynq |
| Observability | OpenTelemetry (traces + metrics) |
| Dashboard | Angular 19 (Zoneless, Signals, Tailwind CSS) |
| Cloud SDKs | aws-sdk-go-v2, azure-sdk-for-go, cloud.google.com/go |
| CI/CD | GitHub Actions, golangci-lint, buf |
| Containers | Docker, Docker Compose |

## Project Structure

```
raftweave/
├── api/proto/               # Protobuf v3 service definitions
├── cmd/
│   ├── raftweave-server/    # Server entry point
│   └── raftweave-cli/       # CLI entry point (Cobra)
├── dashboard/               # Angular 19 dashboard
├── deploy/                  # Docker Compose, Dockerfile, workload schema
├── docs/                    # Architecture & development docs
├── internal/
│   ├── auth/                # OAuth2 (GitHub, Google)
│   ├── build/               # Container build pipeline
│   ├── consensus/           # Custom Raft implementation
│   ├── gen/                 # Generated protobuf Go code
│   ├── health/              # Multi-protocol health probes
│   ├── ingestion/           # Workload submission & management
│   ├── observability/       # OTel tracing & metrics
│   ├── provisioner/         # Multi-cloud provisioning engine
│   │   └── adapters/        # AWS, Azure, GCP adapters
│   ├── replication/         # WAL streaming & standby promotion
│   └── store/               # PostgreSQL data access layer
└── .github/                 # CI workflows, templates
```

## Quick Start

### Prerequisites

- Go 1.24+
- Node.js 22+
- Docker & Docker Compose
- [buf](https://buf.build/docs/installation) CLI

### Local Development

```bash
# Clone the repository
git clone https://github.com/raftweave/raftweave.git
cd raftweave

# Copy environment config
cp deploy/.env.example deploy/.env

# Start full environment (Postgres, Redis, server, dashboard)
make dev

# Or run individual components:
make build        # Build Go binaries
make test         # Run tests with race detector
make lint         # Run golangci-lint
make proto        # Regenerate protobuf code
make dashboard    # Build Angular dashboard
make raft-cluster # Start 3-node Raft cluster
```

### Kubernetes Deployment

RaftWeave can also be deployed to a local or remote Kubernetes cluster using Kustomize.

**Prerequisites:**
- A running Kubernetes cluster (e.g., `minikube start` or Docker Desktop Kubernetes)
- `kubectl` installed

**Deploying the services:**
```bash
# Apply the base kustomization (includes auth, ingestion, build, replication, provisioner, and dashboard)
kubectl apply -k deploy/kubernetes/base

# Check the status of the pods
kubectl get pods -n raftweave
```

### Testing

**Unit and Integration Tests**
Run all Go unit and integration tests (requires Docker for Postgres/Redis):
```bash
make test
```

**End-to-End Tests**
To run the full end-to-end integration tests for ingestion, build, and observability flows:
```bash
go test -v -tags=integration ./tests/integration
```

### Workload Deployment

Define a workload using the declarative YAML format:9

```yaml
apiVersion: raftweave/v1
kind: Workload
metadata:
  name: my-web-app
spec:
  source:
    repository: https://github.com/example-org/my-web-app
  resources:
    cpu: "1.0"
    memoryMB: 512
  primary:
    cloud: aws
    region: us-east-1
  failover:
    cloud: azure
    region: eastus
    automatic: true
```

Apply it with the CLI:

```bash
raftweave-cli apply -f workload.yaml
raftweave-cli status my-web-app
```

See [deploy/workload.example.yaml](deploy/workload.example.yaml) for a complete example with all options.

## Documentation

- [Architecture Overview](docs/ARCHITECTURE.md) — System design, Raft protocol, cloud adapter pattern
- [Local Development Guide](docs/LOCAL_DEVELOPMENT.md) — Setup, debugging, common workflows
- [Contributing](CONTRIBUTING.md) — Code style, branching strategy, PR process

## Development Commands

| Command | Description |
|---------|-------------|
| `make dev` | Start full local environment via Docker Compose |
| `make build` | Build server and CLI binaries |
| `make test` | Run all tests with race detector |
| `make lint` | Run golangci-lint |
| `make proto` | Regenerate protobuf Go code |
| `make dashboard` | Build Angular dashboard for production |
| `make raft-cluster` | Start 3-node Raft cluster |
| `make ci` | Run full CI pipeline locally |
| `make clean` | Remove all build artifacts |

## License

[MIT](LICENSE)

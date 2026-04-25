# Architecture Overview

## System Design

RaftWeave is a sovereign multi-cloud orchestration control plane. It uses a custom Raft consensus implementation to coordinate workload failover across AWS, Azure, and GCP — without relying on any external Raft library.

### Design Principles

- **Clean Architecture**: `domain → usecase → adapter → framework`. Internal packages follow strict layering.
- **No Cross-Package Imports**: Subsystems (`ingestion`, `consensus`, `provisioner`, etc.) communicate only through well-defined interfaces, never through direct imports.
- **Custom Raft**: The consensus layer is built from scratch in pure Go, implementing leader election, log replication, and membership changes via the Raft protocol.
- **Cloud Abstraction**: All cloud-specific logic lives in `internal/provisioner/adapters/`. Each cloud provider implements a common `CloudAdapter` interface.
- **Context Propagation**: Every function that performs I/O accepts `context.Context` as its first parameter.
- **Observable by Default**: All handler methods include OpenTelemetry spans. Metrics are exposed via OTLP.

## Component Architecture

```
                    ┌─────────────────┐
                    │  Connect-RPC    │
                    │   Handlers      │
                    └────────┬────────┘
                             │
           ┌─────────────────┼──────────────────┐
           ▼                 ▼                  ▼
    ┌──────────────┐  ┌─────────────┐  ┌──────────────┐
    │  Ingestion   │  │  Consensus  │  │ Provisioner  │
    │              │  │   (Raft)    │  │              │
    │ - Submit     │  │ - Election  │  │ - Provision  │
    │ - Validate   │  │ - LogRepl   │  │ - Failover   │
    │ - Queue      │  │ - State     │  │ - Fencing    │
    └──────┬───────┘  └──────┬──────┘  └──────┬───────┘
           │                 │                 │
           ▼                 ▼                 ▼
    ┌──────────────┐  ┌─────────────┐  ┌──────────────┐
    │    Build     │  │ Replication │  │   Health     │
    │              │  │             │  │              │
    │ - Detect     │  │ - WAL       │  │ - HTTP       │
    │ - Build      │  │ - Monitor   │  │ - TCP        │
    │ - Push       │  │ - Promote   │  │ - DB         │
    └──────────────┘  └─────────────┘  └──────────────┘
           │                 │                 │
           └─────────────────┼─────────────────┘
                             ▼
                    ┌─────────────────┐
                    │    Store (pgx)  │
                    │    Queue (Redis)│
                    └─────────────────┘
```

## Raft Consensus Protocol

The custom Raft implementation in `internal/consensus/` includes:

### Leader Election
- Nodes start as Followers with randomized election timeouts (150–300ms).
- On timeout, a Follower becomes a Candidate and requests votes.
- A Candidate receiving a majority of votes becomes the Leader.
- The Leader sends periodic heartbeats to maintain authority.

### Log Replication
- The Leader appends client commands as log entries.
- Entries are replicated to Followers via `AppendEntries` RPCs.
- Once a majority acknowledges, the entry is committed and applied to the state machine.

### State Machine
- The Raft log drives a state machine that tracks workload state across the cluster.
- Committed entries trigger provisioning, failover, and replication operations.

### Key Types
- `NodeID`, `Term`, `LogIndex` — Raft protocol primitives
- `RaftRole` — `Leader`, `Follower`, `Candidate`
- `LogEntry` — command entry with term, index, and data
- `Raft` — core engine wiring Election, Log, Transport, StateMachine, and MembershipManager

## Cloud Adapter Pattern

The `internal/provisioner/adapters/` package defines a `CloudAdapter` interface:

```go
type CloudAdapter interface {
    ProvisionCompute(ctx context.Context, req ComputeRequest) (*ComputeResult, error)
    ProvisionDatabase(ctx context.Context, req DatabaseRequest) (*DatabaseResult, error)
    ConfigureDNS(ctx context.Context, req DNSRequest) (*DNSResult, error)
    Deprovision(ctx context.Context, resourceID string) error
    HealthCheck(ctx context.Context, resourceID string) (*HealthResult, error)
}
```

Implementations:
- **AWS** (`adapters/aws/`) — ECS, RDS, Route53
- **Azure** (`adapters/azure/`) — Container Instances, PostgreSQL Flexible Server
- **GCP** (`adapters/gcp/`) — Cloud Run, Cloud SQL

## API Layer

All services use Connect-RPC (compatible with gRPC and gRPC-Web):

| Service | RPCs |
|---------|------|
| IngestionService | SubmitWorkload, AddCloudCredentials, GetWorkloadStatus, ListWorkloads |
| BuildService | TriggerBuild, StreamBuildLogs (server stream), GetBuildResult |
| ConsensusService | GetClusterState, StreamClusterState (server stream), RequestVote, AppendEntries |
| ProvisionerService | ProvisionWorkload, ExecuteFailover, GetResourceStatus |
| ReplicationService | GetReplicationStatus, StreamReplicationMetrics (server stream), PromoteStandby |
| DashboardService | GetOverview, StreamOverview (server stream), GetFailoverLog |
| HealthService | Check, Watch (server stream) |

## Data Model

The PostgreSQL schema (managed via SQL migrations in `internal/store/migrations/`) has five core tables:

1. **workloads** — Workload definitions, status, cloud targets
2. **credentials** — Encrypted cloud provider credentials
3. **build_jobs** — Container build job tracking
4. **raft_state** — Persisted Raft node state (term, voted_for, log)
5. **failover_events** — Audit log of failover operations with RTO/RPO metrics

## Observability

- **Tracing**: Every Connect-RPC handler creates an OTel span. The `observability.TracingInterceptor` adds spans automatically.
- **Metrics**: Seven core metrics are defined in `internal/observability/metrics.go`:
  - `raftweave.workloads.total`
  - `raftweave.builds.duration`
  - `raftweave.builds.total`
  - `raftweave.consensus.term`
  - `raftweave.consensus.leader_changes`
  - `raftweave.replication.lag_seconds`
  - `raftweave.failovers.total`

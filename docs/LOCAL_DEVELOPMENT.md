# Local Development Guide

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.24+ | Backend services |
| Node.js | 22+ | Angular dashboard |
| Docker | Latest | Container runtime |
| Docker Compose | v2+ | Local environment orchestration |
| buf | Latest | Protobuf linting & code generation |
| golangci-lint | v1.64+ | Go linting |

## Initial Setup

### 1. Clone & Configure

```bash
git clone https://github.com/raftweave/raftweave.git
cd raftweave
cp deploy/.env.example deploy/.env
```

Edit `deploy/.env` to set your local configuration. The defaults work out of the box for most setups.

### 2. Install Dependencies

**Go modules:**
```bash
go mod download
```

**Dashboard:**
```bash
cd dashboard
npm install
cd ..
```

**Protobuf tooling:**
```bash
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
```

## Running Locally

### Full Environment (Recommended)

Start all services with Docker Compose:

```bash
make dev
```

This starts:
- **PostgreSQL 16** on port 5432
- **Redis 7** on port 6379
- **Adminer** (DB UI) on port 8081
- **Container Registry** on port 5001
- **RaftWeave Server** on port 8080
- **Angular Dashboard** on port 4200

### Individual Services

**Go server only:**
```bash
go run ./cmd/raftweave-server
```

**Dashboard dev server:**
```bash
cd dashboard
npm start
# Opens at http://localhost:4200
```

**3-Node Raft Cluster:**
```bash
make raft-cluster
```

This starts three RaftWeave nodes that form a consensus cluster.

## Development Workflow

### Making Changes

1. Create a feature branch from `develop`
2. Make your changes
3. Run the verification suite:

```bash
# Format Go code
gofmt -w cmd/ internal/

# Lint
make lint

# Test
make test

# Build
make build
```

### Protobuf Changes

When modifying `.proto` files:

```bash
# Lint proto files
buf lint api/proto

# Regenerate Go code
make proto

# Verify generated code compiles
go build ./...
```

### Dashboard Changes

```bash
cd dashboard

# Dev server with hot reload
npm start

# Production build
npx ng build --configuration=production

# Lint (requires @angular-eslint setup)
npx ng lint
```

## Database

### Migrations

SQL migrations are in `internal/store/migrations/`. Each migration has an `up` and `down` file:

```
001_workloads.up.sql / 001_workloads.down.sql
002_credentials.up.sql / 002_credentials.down.sql
003_build_jobs.up.sql / 003_build_jobs.down.sql
004_raft_state.up.sql / 004_raft_state.down.sql
005_failover_events.up.sql / 005_failover_events.down.sql
```

### Connecting to the Database

With the Docker environment running:

```bash
# Via psql
psql -h localhost -p 5432 -U raftweave -d raftweave

# Via Adminer web UI
open http://localhost:8081
```

Default credentials (from `.env.example`):
- Host: `localhost`
- Port: `5432`
- Database: `raftweave`
- User: `raftweave`
- Password: `raftweave_dev`

## Docker

### Rebuilding Images

```bash
# Rebuild and restart
docker compose -f deploy/docker-compose.yml up --build

# Rebuild a specific service
docker compose -f deploy/docker-compose.yml build raftweave-server
```

### Cleaning Up

```bash
# Stop and remove containers
docker compose -f deploy/docker-compose.yml down

# Remove volumes (resets database)
docker compose -f deploy/docker-compose.yml down -v

# Remove all build artifacts
make clean
```

## Environment Variables

Key environment variables (see `deploy/.env.example` for full list):

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Server listen port |
| `DATABASE_URL` | `postgres://raftweave:raftweave_dev@localhost:5432/raftweave` | PostgreSQL connection |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection |
| `NODE_ID` | `node-1` | Raft node identifier |
| `CLUSTER_PEERS` | *(empty)* | Comma-separated peer addresses |
| `OTEL_EXPORTER_ENDPOINT` | `localhost:4317` | OTel collector endpoint |
| `GITHUB_CLIENT_ID` | *(empty)* | GitHub OAuth app ID |
| `GOOGLE_CLIENT_ID` | *(empty)* | Google OAuth app ID |

## Troubleshooting

### Go build fails
```bash
# Ensure workspace mode is working
go work sync
go mod tidy
go build ./...
```

### Proto generation fails
```bash
# Verify buf and protoc plugins are installed
buf --version
which protoc-gen-go
which protoc-gen-connect-go

# Regenerate from api/proto directory
cd api/proto && buf generate
```

### Dashboard build fails
```bash
# Clear Angular cache
rm -rf dashboard/.angular
cd dashboard && npm ci
npx ng build --configuration=production
```

### Docker container won't start
```bash
# Check logs
docker compose -f deploy/docker-compose.yml logs raftweave-server

# Verify environment
docker compose -f deploy/docker-compose.yml config
```

.PHONY: dev proto test lint build dashboard ci raft-cluster clean

# Start full local development environment
dev:
	docker compose -f deploy/docker-compose.yml up --build

# Generate Go code from proto files
proto:
	buf generate api/proto

# Run all Go tests with race detector
test:
	go test -race -count=1 ./...

# Run golangci-lint across all modules
lint:
	golangci-lint run ./...

# Build all Go binaries
build:
	go build -o bin/raftweave-server ./cmd/raftweave-server
	go build -o bin/raftweave-cli ./cmd/raftweave-cli

# Build Angular dashboard
dashboard:
	cd dashboard && npm ci && npx ng build --configuration=production

# Run full CI pipeline locally
ci: proto lint test build

# Start 3-node Raft cluster
raft-cluster:
	docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.raft.yml up --build raftweave-node-1 raftweave-node-2 raftweave-node-3

# Remove all build artifacts
clean:
	rm -rf bin/ dist/ tmp/ coverage/
	rm -rf dashboard/dist/ dashboard/.angular/
	rm -rf internal/gen/

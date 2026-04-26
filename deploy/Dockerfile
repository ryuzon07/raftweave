FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY go.mod go.sum ./
COPY cmd/raftweave-cli/go.mod cmd/raftweave-cli/go.sum ./cmd/raftweave-cli/

# Download dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build the binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /raftweave-server ./cmd/raftweave-server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
COPY --from=builder /raftweave-server /usr/local/bin/raftweave-server

ENTRYPOINT ["raftweave-server"]

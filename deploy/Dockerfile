FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY go.mod go.sum ./
COPY cmd/ cmd/
COPY internal/ internal/
COPY api/ api/

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /raftweave-server ./cmd/raftweave-server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
COPY --from=builder /raftweave-server /usr/local/bin/raftweave-server

ENTRYPOINT ["raftweave-server"]

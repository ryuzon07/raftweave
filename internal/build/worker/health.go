package worker

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthHandler manages the worker's liveness and readiness logic.
type HealthHandler struct {
	pgPool *pgxpool.Pool
	redis  *redis.Client
}

// NewHealthHandler creates a new instance of HealthHandler.
func NewHealthHandler(pgPool *pgxpool.Pool, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		pgPool: pgPool,
		redis:  redisClient,
	}
}

// Liveness provides a fast HTTP 200 indicating the binary is running.
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Readiness performs health checks on external dependency connections.
// Returns HTTP 503 if any dependencies are unreachable.
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check postgres
	if err := h.pgPool.Ping(ctx); err != nil {
		http.Error(w, "PostgreSQL unavailable", http.StatusServiceUnavailable)
		return
	}

	// Check Redis
	res := h.redis.Ping(context.Background())
	if err := res.Err(); err != nil {
		http.Error(w, "Redis unavailable", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("READY"))
}

// RegisterHealthEndpoints sets up the standard `healthz` and `readyz`
// endpoints on a provided HTTP multiplexer.
func RegisterHealthEndpoints(mux *http.ServeMux, handler *HealthHandler) {
	mux.HandleFunc("/healthz", handler.Liveness)
	mux.HandleFunc("/readyz", handler.Readiness)
}

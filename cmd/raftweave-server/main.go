package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	authdi "github.com/raftweave/raftweave/internal/auth/di"
	ingestiondi "github.com/raftweave/raftweave/internal/ingestion/di"
)

func main() {
	fmt.Println("--- RAFTWEAVE VERSION 1.0.1 STARTING ---")
	mode := flag.String("mode", "ingestion", "Service mode (ingestion, auth, build, consensus, replication, provisioner)")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("starting raftweave server v13", zap.String("mode", *mode))

	// Common dependencies
	dbPool := initDB(ctx, logger)
	defer dbPool.Close()

	rdb := initRedis(ctx, logger)
	defer rdb.Close()

	var handler http.Handler
	var port string

	switch *mode {
	case "ingestion":
		handler, port = startIngestion(ctx, dbPool, rdb, logger)
	case "auth":
		handler, port = startAuth(ctx, dbPool, rdb, logger)
	case "build", "consensus", "replication", "provisioner":
		handler, port = startStub(ctx, *mode, logger)
	default:
		logger.Fatal("unsupported mode", zap.String("mode", *mode))
	}

	// Wrap with aggressive CORS at the very top level
	finalHandler := corsMiddleware(handler, logger)

	server := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: finalHandler,
	}

	go func() {
		fmt.Printf("\n\n##########################################\n")
		fmt.Printf("!!! RAFTWEAVE AUTH SERVER v13 STARTING !!!\n")
		fmt.Printf("##########################################\n\n\n")
		logger.Info("server listening", zap.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}
}

func initDB(ctx context.Context, logger *zap.Logger) *pgxpool.Pool {
	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	db := os.Getenv("POSTGRES_DB")
	port := os.Getenv("POSTGRES_PORT")
	if port == "" { port = "5432" }

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Fatal("failed to parse postgres dsn", zap.Error(err))
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	return pool
}

func initRedis(ctx context.Context, logger *zap.Logger) *redis.Client {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	pass := os.Getenv("REDIS_PASSWORD")
	if port == "" { port = "6379" }

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: pass,
		DB:       0,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatal("failed to connect to redis", zap.Error(err))
	}
	return rdb
}

func startIngestion(ctx context.Context, db *pgxpool.Pool, rdb *redis.Client, logger *zap.Logger) (http.Handler, string) {
	cfg := ingestiondi.Config{
		DBPool: db,
		// Ingestion uses asynq client opt
		Redis: asynq.RedisClientOpt{
			Addr:     rdb.Options().Addr,
			Password: rdb.Options().Password,
			DB:       rdb.Options().DB,
		},
		Logger: logger,
		CryptoKeys: map[string][]byte{
			"v1": []byte(os.Getenv("ENCRYPTION_KEY_V1")),
		},
		ActiveCryptoVersion: "v1",
	}

	mod, err := ingestiondi.Bootstrap(ctx, cfg)
	if err != nil {
		logger.Fatal("failed to bootstrap ingestion", zap.Error(err))
	}

	mux := http.NewServeMux()
	// Connect-RPC RPCs
	mux.Handle("/api"+mod.RPCHandlerPath, http.StripPrefix("/api", mod.RPCHandler))
	mux.Handle(mod.RPCHandlerPath, mod.RPCHandler)

	mux.Handle("/v1/webhooks/ingest", mod.WebhookHandler)
	
	// Add a health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return mux, "8080"
}

func startAuth(ctx context.Context, db *pgxpool.Pool, rdb *redis.Client, logger *zap.Logger) (http.Handler, string) {
	jwtKey := os.Getenv("JWT_PRIVATE_KEY_PEM")
	jwtKey = strings.ReplaceAll(jwtKey, "\r", "")

	cfg := authdi.Config{
		DBPool:             db,
		Redis:              rdb,
		Logger:             logger,
		JWTPrivateKeyPEM:   []byte(jwtKey),
		EncryptionKey:      []byte(os.Getenv("ENCRYPTION_KEY")),
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		CookieDomain:       os.Getenv("COOKIE_DOMAIN"),
		DashboardURL:       os.Getenv("DASHBOARD_URL"),
	}

	mod, err := authdi.Bootstrap(ctx, cfg)
	if err != nil {
		logger.Fatal("failed to bootstrap auth", zap.Error(err))
	}

	mux := http.NewServeMux()
	// Connect-RPC RPCs
	mux.Handle("/api/auth.v1.AuthService/", http.StripPrefix("/api", mod.RPCHandler))
	mux.Handle("/auth.v1.AuthService/", mod.RPCHandler)

	// OAuth login and callback flows
	mux.Handle("/api/auth/", http.StripPrefix("/api", mod.OAuthHandler))
	mux.Handle("/auth/", mod.OAuthHandler)

	// Add health checks
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	// Wrap with CORS middleware
	handler := corsMiddleware(mux, logger)

	return handler, "8080"
}

func corsMiddleware(next http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ULTIMATE DEBUG LOG
		fmt.Printf(">>> [HTTP] %s %s (Origin: %s)\n", r.Method, r.URL.Path, r.Header.Get("Origin"))

		// Aggressively set headers for EVERY request
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Connect-Protocol-Version, Connect-Timeout-Ms")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight immediately
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func startStub(ctx context.Context, mode string, logger *zap.Logger) (http.Handler, string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintf(w, "Service %s is not yet fully implemented in this version.", mode)
	})
	return mux, "8080"
}

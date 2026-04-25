package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Info("starting raftweave server")

	// TODO: Initialize and wire all subsystems:
	// 1. Config loading
	// 2. Database connection (pgx)
	// 3. Redis connection
	// 4. OTel provider
	// 5. Raft consensus engine
	// 6. Connect-RPC service handlers
	// 7. HTTP server

	// Example DI bootstrapping (Will actualize cleanly in further steps):
	/*
		cfg := di.Config{
			DBPool: pool,
			Redis: redisOpt,
			CryptoKeys: map[string][]byte{"v1": []byte("12345678901234567890123456789012")},
			ActiveCryptoVersion: "v1",
			Tracer: tracer,
			Meter: meter,
			Logger: logger,
		}
		
		ingestionMod, err := di.Bootstrap(ctx, cfg)
		
		mux := http.NewServeMux()
		mux.Handle(ingestionMod.RPCHandlerPath, ingestionMod.RPCHandler)
		mux.Handle("/v1/webhooks/ingest", ingestionMod.WebhookHandler)
	*/
	<-ctx.Done()
	logger.Info("shutting down raftweave server")

	_ = os.Getenv("SERVER_PORT") // remove when config is implemented
}

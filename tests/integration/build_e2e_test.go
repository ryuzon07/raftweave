//go:build integration
// +build integration

package integration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel/metric/noop"
	traceNoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"github.com/raftweave/raftweave/internal/build/adapter/detector"
	"github.com/raftweave/raftweave/internal/build/adapter/dockerfile"
	"github.com/raftweave/raftweave/internal/build/adapter/logstream"
	"github.com/raftweave/raftweave/internal/build/adapter/postgres"
	"github.com/raftweave/raftweave/internal/build/domain"
	buildmetrics "github.com/raftweave/raftweave/internal/build/metrics"
	"github.com/raftweave/raftweave/internal/build/usecase"
	"github.com/raftweave/raftweave/internal/build/worker"
)

// ---------------------------------------------------------------------------
// Test Infrastructure Setup
// ---------------------------------------------------------------------------

func setupBuildTestPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		tcpostgres.WithDatabase("raftweave_build_e2e"),
		tcpostgres.WithUsername("raftweave"),
		tcpostgres.WithPassword("e2e_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)

	if err != nil {
		t.Skipf("failed to start postgres container (Docker may not be available on Windows): %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	// Create required schema for tests here (normally handled by Goose during startup)
	_, err = pool.Exec(ctx, `
		CREATE TABLE builds (
			id TEXT PRIMARY KEY,
			workload_id TEXT NOT NULL,
			workspace_id TEXT NOT NULL,
			git_commit_sha TEXT NOT NULL,
			git_branch TEXT NOT NULL,
			status TEXT NOT NULL,
			language TEXT,
			image_ref TEXT,
			image_digest TEXT,
			size_bytes BIGINT,
			error_message TEXT,
			started_at TIMESTAMP WITH TIME ZONE,
			completed_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
		CREATE TABLE build_log_lines (
			build_id TEXT NOT NULL REFERENCES builds(id),
			sequence BIGINT NOT NULL,
			stream TEXT NOT NULL,
			text TEXT NOT NULL,
			ts TIMESTAMP WITH TIME ZONE NOT NULL,
			PRIMARY KEY (build_id, sequence)
		);
	`)
	require.NoError(t, err)

	return pool
}

func setupBuildTestRedis(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	container, err := tcredis.RunContainer(ctx,
		testcontainers.WithImage("docker.io/redis:7-alpine"),
	)
	if err != nil {
		t.Skipf("failed to start redis container (Docker may not be available on Windows): %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	return connStr
}

// ---------------------------------------------------------------------------
// End-to-End Test for the Build Queue & Worker Pipeline
// ---------------------------------------------------------------------------

func TestBuildWorker_E2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	logger, _ := zap.NewDevelopment()
	tracer := traceNoop.NewTracerProvider().Tracer("test")
	meter := noop.NewMeterProvider().Meter("test")
	_ = buildmetrics.InitMetrics(meter)

	// 1. Setup Postgres and Redis
	dbPool := setupBuildTestPostgres(t)
	redisAddr := setupBuildTestRedis(t)

	// 2. Setup Build Repository and Usecase logic
	repo := postgres.NewBuildRepository(dbPool, tracer)
	logsRepo := postgres.NewLogRepository(dbPool, tracer)
	sysDetector := detector.New()
	sysGenerator := dockerfile.New()
	sysBroadcaster := logstream.NewRedis(redisAddr)

	// We'll pass nil or dummy for kaniko and pusher in tests since it requires real docker socket in cluster
	useCase := usecase.New(repo, logsRepo, sysDetector, sysGenerator, nil, nil, sysBroadcaster, tracer)

	// 3. Setup Asynq Worker and Client
	redisOpt, err := asynq.ParseRedisURI(redisAddr)
	require.NoError(t, err)

	queueClient := asynq.NewClient(redisOpt)
	defer queueClient.Close()

	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 1,
			Queues: map[string]int{
				"builds:default": 10,
			},
		},
	)

	mux := asynq.NewServeMux()
	handler := worker.NewHandler(useCase)
	mux.Handle(worker.TaskTypeBuild, handler)

	// Run Worker in background
	go func() {
		if err := srv.Run(mux); err != nil {
			logger.Error("asynq server error", zap.Error(err))
		}
	}()
	defer srv.Stop()

	// 4. Create and Enqueue a Build Record
	workloadID := "test-w-12345"
	job := usecase.BuildJob{
		WorkloadID:   workloadID,
		WorkspaceID:  "workspace-999",
		GitCommitSHA: "abc123def",
		GitRepoURL:   "https://github.com/raftweave/example",
		GitBranch:    "main",
		SourcePath:   "/tmp/src",
	}

	payload, err := json.Marshal(job)
	require.NoError(t, err)

	task := asynq.NewTask(worker.TaskTypeBuild, payload)

	_, err = queueClient.Enqueue(task, asynq.Queue("builds:default"))
	require.NoError(t, err)

	// 5. Poll the Database to detect creation and failure state
	// It should create a build internally and then fail when Kaniko is nil,
	// or fail when detector executes without real files at /tmp/src.

	require.Eventually(t, func() bool {
		builds, fetchErr := repo.ListByWorkload(ctx, workloadID, 10, 0)
		if fetchErr != nil || len(builds) == 0 {
			return false
		}

		status := builds[0].Status
		return status == domain.BuildStatusFailed || status == domain.BuildStatusSucceeded
	}, 15*time.Second, 500*time.Millisecond, "Build status should transition to completed or failed")

	builds, err := repo.ListByWorkload(ctx, workloadID, 10, 0)
	require.NoError(t, err)
	require.Len(t, builds, 1)
	require.Equal(t, domain.BuildStatusFailed, builds[0].Status, "Expected Build Status Failed due to lack of kaniko/repo files")

	// Ensure log tracking worked
	lines, _ := logsRepo.GetLines(ctx, builds[0].ID, 0)
	require.Greater(t, len(lines), 0, "Build logs should contain state machine records")
}

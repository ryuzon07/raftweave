//go:build integration
// +build integration

package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	v1 "github.com/raftweave/raftweave/internal/gen/raftweave/v1"
	"github.com/raftweave/raftweave/internal/gen/raftweave/v1/raftweavev1connect"
	"github.com/raftweave/raftweave/internal/ingestion"
	"github.com/raftweave/raftweave/internal/ingestion/adapter/crypto"
	pgadapter "github.com/raftweave/raftweave/internal/ingestion/adapter/postgres"
	"github.com/raftweave/raftweave/internal/ingestion/adapter/queue"
	"github.com/raftweave/raftweave/internal/ingestion/handler"
	"github.com/raftweave/raftweave/internal/ingestion/usecase"
)

// TestObservability verifies that all required spans, metrics, and structured logs are present
// during a SubmitWorkload flow.
func TestObservability_SubmitWorkload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping observability test in short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. Setup OTel Trace Exporter
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := trace.NewTracerProvider(trace.WithSpanProcessor(spanRecorder))
	tracer := tracerProvider.Tracer("observability-test")

	// 2. Setup OTel Metric Exporter
	metricReader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(metric.WithReader(metricReader))
	meter := meterProvider.Meter("observability-test")
	metrics := ingestion.NewIngestionMetrics(meter)

	// 3. Setup Zap Observer
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core).With(zap.String("component", "ingestion"))

	// Infrastructure (reuse from E2E)
	pgPool := setupTestPostgres(t)
	redisOpt := setupTestRedis(t)

	encryptor, err := crypto.NewAESEncryptor(map[string][]byte{"v1": []byte("01234567890123456789012345678901")}, "v1")
	require.NoError(t, err)

	workloadRepo := pgadapter.NewWorkloadRepository(pgPool, tracer)
	credentialRepo := pgadapter.NewCredentialRepository(pgPool, tracer)
	enqueuer := queue.NewAsynqJobEnqueuer(redisOpt, tracer, logger)

	deps := usecase.Dependencies{
		WorkloadRepo:   workloadRepo,
		CredentialRepo: credentialRepo,
		JobEnqueuer:    enqueuer,
		Encryptor:      encryptor,
		Tracer:         tracer,
		Logger:         logger,
		Metrics:        metrics,
	}

	submitWorkloadUC := usecase.NewSubmitWorkloadUseCase(deps)
	addCredentialUC := usecase.NewAddCredentialUseCase(deps)

	ingestionHandler := handler.NewIngestionHandler(logger, submitWorkloadUC, addCredentialUC, workloadRepo)

	mux := http.NewServeMux()
	path, h := raftweavev1connect.NewIngestionServiceHandler(ingestionHandler)
	mux.Handle(path, h)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := raftweavev1connect.NewIngestionServiceClient(srv.Client(), srv.URL)

	// Trigger the request
	req := &v1.SubmitWorkloadRequest{
		Workload: &v1.WorkloadDescriptor{
			Name: "obs-workload",
			Source: &v1.SourceSpec{Type: "git", RepoUrl: "https://github.com/obs/repo", Branch: "main"},
			Regions: &v1.RegionConfig{
				Primary: &v1.RegionTarget{Provider: v1.CloudProvider_CLOUD_PROVIDER_AWS, Region: v1.Region_REGION_AWS_AP_SOUTH_1},
				Standbys: []*v1.RegionTarget{{Provider: v1.CloudProvider_CLOUD_PROVIDER_AWS, Region: v1.Region_REGION_AWS_US_EAST_1}},
			},
			Compute: &v1.ResourceSpec{Cpu: "1", Memory: "1Gi", Replicas: 1},
			Database: &v1.DatabaseSpec{Engine: "postgres", Version: "16", StorageGb: 10},
			Failover: &v1.FailoverConfig{RtoSeconds: 30, RpoSeconds: 5, AutoFailover: true, MinHealthyNodes: 1, FencingEnabled: true},
		},
	}
	_, err = client.SubmitWorkload(ctx, connect.NewRequest(req))
	require.NoError(t, err)

	// Force flush traces
	err = tracerProvider.ForceFlush(ctx)
	require.NoError(t, err)

	// --- A. Verify Spans ---
	spans := spanRecorder.Ended()
	spanNames := make([]string, len(spans))
	for i, s := range spans {
		spanNames[i] = s.Name()
	}

	assert.Contains(t, spanNames, "ingestion.SubmitWorkload")
	assert.Contains(t, spanNames, "postgres.WorkloadRepository.FindByName")
	assert.Contains(t, spanNames, "postgres.WorkloadRepository.Save")
	assert.Contains(t, spanNames, "queue.EnqueueProvisionJob")

	// --- B. Verify Metrics ---
	rm := metricdata.ResourceMetrics{}
	err = metricReader.Collect(ctx, &rm)
	require.NoError(t, err)

	var foundWorkloadsSubmitted bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "raftweave.ingestion.workloads_submitted_total" {
				foundWorkloadsSubmitted = true
				data, ok := m.Data.(metricdata.Sum[int64])
				require.True(t, ok)
				assert.Len(t, data.DataPoints, 1)
				assert.Equal(t, int64(1), data.DataPoints[0].Value)
			}
		}
	}
	assert.True(t, foundWorkloadsSubmitted, "metric workloads_submitted_total should be recorded")
	// Since we only called SubmitWorkload, webhook duration might not be populated in output,
	// but the metric should be registered. Wait, metricReader.Collect only returns metrics that have data points.
	// So we won't assert foundWebhookDuration here unless we trigger processWebhook.

	// --- C. Verify Structured Logs ---
	logEntries := logs.All()
	var foundSubmitLog bool
	for _, entry := range logEntries {
		if entry.Message == "workload_submitted" {
			foundSubmitLog = true
			contextMap := entry.ContextMap()
			assert.Contains(t, contextMap, "workload_id")
			assert.Contains(t, contextMap, "workload_name")
			assert.Contains(t, contextMap, "primary_region")
			assert.Equal(t, "ingestion", contextMap["component"])
		}
	}
	assert.True(t, foundSubmitLog, "should find structured log for workload_submitted")
}

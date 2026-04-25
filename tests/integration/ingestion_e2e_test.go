//go:build integration
// +build integration

package integration_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel/metric/noop"
	traceNoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	v1 "github.com/raftweave/raftweave/internal/gen/raftweave/v1"
	"github.com/raftweave/raftweave/internal/gen/raftweave/v1/raftweavev1connect"
	"github.com/raftweave/raftweave/internal/ingestion"
	"github.com/raftweave/raftweave/internal/ingestion/adapter/crypto"
	pgadapter "github.com/raftweave/raftweave/internal/ingestion/adapter/postgres"
	"github.com/raftweave/raftweave/internal/ingestion/adapter/queue"
	"github.com/raftweave/raftweave/internal/ingestion/domain"
	"github.com/raftweave/raftweave/internal/ingestion/handler"
	"github.com/raftweave/raftweave/internal/ingestion/usecase"
)

// ---------------------------------------------------------------------------
// Test Infrastructure Setup
// ---------------------------------------------------------------------------

func setupTestPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		tcpostgres.WithDatabase("raftweave_e2e"),
		tcpostgres.WithUsername("raftweave"),
		tcpostgres.WithPassword("e2e_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Apply schema
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	schema := `
	CREATE TABLE IF NOT EXISTS workloads (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		descriptor_json JSONB NOT NULL,
		status VARCHAR(50) NOT NULL,
		primary_region VARCHAR(100) NOT NULL,
		primary_provider VARCHAR(50) NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);
	CREATE TABLE IF NOT EXISTS credentials (
		id VARCHAR(36) PRIMARY KEY,
		workload_id VARCHAR(36) NOT NULL REFERENCES workloads(id) ON DELETE CASCADE,
		provider VARCHAR(50) NOT NULL,
		credential_type VARCHAR(50) NOT NULL,
		encrypted_payload BYTEA NOT NULL,
		key_version VARCHAR(50) NOT NULL,
		created_at TIMESTAMP NOT NULL,
		rotated_at TIMESTAMP
	);`
	_, err = pool.Exec(ctx, schema)
	require.NoError(t, err)

	return pool
}

func setupTestRedis(t *testing.T) asynq.RedisClientOpt {
	t.Helper()
	ctx := context.Background()

	redisContainer, err := tcredis.RunContainer(ctx,
		testcontainers.WithImage("docker.io/redis:7-alpine"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections"),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = redisContainer.Terminate(ctx) })

	uri, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	opts, err := asynq.ParseRedisURI(uri)
	require.NoError(t, err)

	return opts.(asynq.RedisClientOpt)
}

// staticWebhookSecretStore implements usecase.WebhookSecretStore for testing.
type staticWebhookSecretStore struct {
	secrets map[string]string
}

func (s *staticWebhookSecretStore) GetSecret(_ context.Context, workloadName string) (string, error) {
	secret, ok := s.secrets[workloadName]
	if !ok {
		return "", fmt.Errorf("no webhook secret for workload %q", workloadName)
	}
	return secret, nil
}

// ---------------------------------------------------------------------------
// E2E Test: Full Ingestion Layer Flow
// ---------------------------------------------------------------------------

// TestIngestionLayer_E2E_GitHubWebhookToJobQueue exercises the complete
// Ingestion Layer flow from HTTP request to Redis job queue using real
// infrastructure via testcontainers.
//
// Scenario:
//  1. Submit a workload descriptor via Connect-RPC
//  2. Add AWS credentials for the workload
//  3. Send a simulated GitHub webhook push event
//  4. Assert: workload exists in PostgreSQL with PENDING status
//  5. Assert: credential exists in PostgreSQL (encrypted, not plaintext)
//  6. Assert: provision job exists in Redis queue (from step 1)
//  7. Assert: build job exists in Redis queue (from step 3)
func TestIngestionLayer_E2E_GitHubWebhookToJobQueue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// ── Infrastructure ────────────────────────────────────────────
	pgPool := setupTestPostgres(t)
	redisOpt := setupTestRedis(t)

	// ── Observability (noop for testing) ──────────────────────────
	tracer := traceNoop.NewTracerProvider().Tracer("e2e-test")
	logger := zap.NewNop()
	meter := noop.NewMeterProvider().Meter("e2e-test")
	metrics := ingestion.NewIngestionMetrics(meter)

	// ── Encryption ────────────────────────────────────────────────
	testKey := []byte("01234567890123456789012345678901") // 32 bytes
	encryptor, err := crypto.NewAESEncryptor(
		map[string][]byte{"v1": testKey},
		"v1",
	)
	require.NoError(t, err)

	// ── Repositories ──────────────────────────────────────────────
	workloadRepo := pgadapter.NewWorkloadRepository(pgPool, tracer)
	credentialRepo := pgadapter.NewCredentialRepository(pgPool, tracer)

	// ── Queue ─────────────────────────────────────────────────────
	enqueuer := queue.NewAsynqJobEnqueuer(redisOpt, tracer, logger)

	// ── Use Cases ─────────────────────────────────────────────────
	deps := usecase.Dependencies{
		WorkloadRepo:   workloadRepo,
		CredentialRepo: credentialRepo,
		JobEnqueuer:    enqueuer,
		Encryptor:      encryptor,
		Tracer:         tracer,
		Logger:         logger,
		Metrics:        metrics,
	}

	webhookSecret := "test-webhook-secret-12345"
	secretStore := &staticWebhookSecretStore{
		secrets: map[string]string{
			"test-workload": webhookSecret,
		},
	}

	submitWorkloadUC := usecase.NewSubmitWorkloadUseCase(deps)
	addCredentialUC := usecase.NewAddCredentialUseCase(deps)
	processWebhookUC := usecase.NewProcessWebhookUseCase(deps, secretStore)

	// ── Handlers ──────────────────────────────────────────────────
	ingestionHandler := handler.NewIngestionHandler(
		logger,
		submitWorkloadUC,
		addCredentialUC,
		workloadRepo,
	)
	webhookHandler := handler.NewWebhookHandler(logger, processWebhookUC)

	// ── HTTP Server ───────────────────────────────────────────────
	mux := http.NewServeMux()
	path, h := raftweavev1connect.NewIngestionServiceHandler(ingestionHandler)
	mux.Handle(path, h)
	mux.Handle("/webhooks/github", webhookHandler)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := raftweavev1connect.NewIngestionServiceClient(
		srv.Client(),
		srv.URL,
	)

	// ══════════════════════════════════════════════════════════════
	// STEP 1: Submit a Workload Descriptor via Connect-RPC
	// ══════════════════════════════════════════════════════════════
	t.Log("Step 1: Submitting workload via Connect-RPC...")

	submitResp, err := client.SubmitWorkload(ctx, connect.NewRequest(&v1.SubmitWorkloadRequest{
		Workload: &v1.WorkloadDescriptor{
			Name: "test-workload",
			Source: &v1.SourceSpec{
				Type:       "git",
				RepoUrl:    "https://github.com/test/repo",
				Branch:     "main",
				Dockerfile: "Dockerfile",
			},
			Regions: &v1.RegionConfig{
				Primary: &v1.RegionTarget{
					Provider: v1.CloudProvider_CLOUD_PROVIDER_AWS,
					Region:   v1.Region_REGION_AWS_AP_SOUTH_1,
				},
				Standbys: []*v1.RegionTarget{
					{
						Provider: v1.CloudProvider_CLOUD_PROVIDER_AWS,
						Region:   v1.Region_REGION_AWS_US_EAST_1,
					},
				},
			},
			Compute: &v1.ResourceSpec{
				Cpu:      "2",
				Memory:   "4Gi",
				Replicas: 2,
			},
			Database: &v1.DatabaseSpec{
				Engine:    "postgres",
				Version:   "16",
				StorageGb: 50,
			},
			Failover: &v1.FailoverConfig{
				RtoSeconds:      30,
				RpoSeconds:      5,
				AutoFailover:    true,
				MinHealthyNodes: 2,
				FencingEnabled:  true,
			},
		},
	}))
	require.NoError(t, err, "SubmitWorkload should succeed")
	require.NotEmpty(t, submitResp.Msg.GetWorkloadId(), "should return a workload ID")
	assert.Equal(t, v1.WorkloadStatus_WORKLOAD_STATUS_PENDING, submitResp.Msg.GetStatus(),
		"new workload should be PENDING")
	t.Logf("  ✓ Workload submitted: ID=%s, Status=%s",
		submitResp.Msg.GetWorkloadId(), submitResp.Msg.GetStatus())

	workloadID := submitResp.Msg.GetWorkloadId()

	// ══════════════════════════════════════════════════════════════
	// STEP 2: Add AWS Credentials for the Workload
	// ══════════════════════════════════════════════════════════════
	t.Log("Step 2: Adding AWS credentials...")

	rawCredential := []byte(`{"access_key":"AKIAIOSFODNN7EXAMPLE","secret_key":"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"}`)

	credResp, err := client.AddCloudCredentials(ctx, connect.NewRequest(&v1.AddCloudCredentialsRequest{
		Credentials: &v1.CloudCredentials{
			Id:               workloadID,
			Provider:         v1.CloudProvider_CLOUD_PROVIDER_AWS,
			CredentialType:   "aws_iam",
			EncryptedPayload: rawCredential, // sent as raw, encrypted by domain
		},
	}))
	require.NoError(t, err, "AddCloudCredentials should succeed")
	require.True(t, credResp.Msg.GetSuccess(), "credential addition should succeed")
	require.NotEmpty(t, credResp.Msg.GetCredentialId(), "should return a credential ID")
	t.Logf("  ✓ Credential added: ID=%s", credResp.Msg.GetCredentialId())

	// ══════════════════════════════════════════════════════════════
	// STEP 3: Send a Simulated GitHub Webhook Push Event
	// ══════════════════════════════════════════════════════════════
	t.Log("Step 3: Sending simulated GitHub webhook...")

	webhookPayload := map[string]interface{}{
		"ref": "refs/heads/main",
		"head_commit": map[string]interface{}{
			"id":      "abc123def456",
			"message": "feat: add new endpoint",
		},
		"repository": map[string]interface{}{
			"clone_url": "https://github.com/test/repo",
		},
		"pusher": map[string]interface{}{
			"name": "testuser",
		},
	}
	payloadBytes, err := json.Marshal(webhookPayload)
	require.NoError(t, err)

	// Compute HMAC-SHA256 signature matching domain.WebhookEvent.VerifySignature
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(payloadBytes)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	webhookReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		srv.URL+"/webhooks/github", bytes.NewReader(payloadBytes))
	require.NoError(t, err)
	webhookReq.Header.Set("Content-Type", "application/json")
	webhookReq.Header.Set("X-GitHub-Event", "push")
	webhookReq.Header.Set("X-Hub-Signature-256", signature)

	webhookResp, err := srv.Client().Do(webhookReq)
	require.NoError(t, err)
	defer webhookResp.Body.Close()

	assert.Equal(t, http.StatusAccepted, webhookResp.StatusCode,
		"webhook should return 202 Accepted")
	t.Logf("  ✓ Webhook accepted: HTTP %d", webhookResp.StatusCode)

	// ══════════════════════════════════════════════════════════════
	// STEP 4: Assert Workload Exists in PostgreSQL with PENDING status
	// ══════════════════════════════════════════════════════════════
	t.Log("Step 4: Verifying workload in PostgreSQL...")

	w, err := workloadRepo.FindByID(ctx, domain.WorkloadID(workloadID))
	require.NoError(t, err, "workload should exist in PostgreSQL")
	assert.Equal(t, "test-workload", w.Name)
	assert.Equal(t, domain.WorkloadStatusPending, w.Status,
		"workload should still be PENDING")
	assert.Equal(t, "ap-south-1", w.PrimaryRegion.Name)
	assert.Equal(t, domain.CloudProviderAWS, w.PrimaryRegion.Provider)
	assert.Len(t, w.StandbyRegions, 1, "should have 1 standby region")
	t.Logf("  ✓ Workload verified: Name=%s, Status=%s, Region=%s",
		w.Name, w.Status, w.PrimaryRegion.Name)

	// ══════════════════════════════════════════════════════════════
	// STEP 5: Assert Credential Exists (Encrypted, Not Plaintext)
	// ══════════════════════════════════════════════════════════════
	t.Log("Step 5: Verifying credential in PostgreSQL...")

	cred, err := credentialRepo.FindByWorkloadAndProvider(ctx,
		domain.WorkloadID(workloadID), domain.CloudProviderAWS)
	require.NoError(t, err, "credential should exist in PostgreSQL")
	assert.Equal(t, domain.CloudProviderAWS, cred.Provider)
	assert.Equal(t, domain.CredentialType("aws_iam"), cred.Type)
	assert.NotEqual(t, rawCredential, cred.EncryptedPayload,
		"stored payload MUST be encrypted, not plaintext")
	assert.Equal(t, "v1", cred.KeyVersion, "key version should be v1")

	// Verify decryption roundtrip
	decrypted, err := cred.Decrypt(encryptor)
	require.NoError(t, err, "decryption should succeed")
	assert.Equal(t, rawCredential, decrypted,
		"decrypted payload should match original raw credential")
	t.Log("  ✓ Credential verified: encrypted at rest, decryption roundtrip OK")

	// ══════════════════════════════════════════════════════════════
	// STEP 6: Assert Provision Job Exists in Redis Queue
	// ══════════════════════════════════════════════════════════════
	t.Log("Step 6: Verifying provision job in Redis queue...")

	// Use asynq Inspector to verify queued tasks
	inspector := asynq.NewInspector(redisOpt)
	defer inspector.Close()

	// Give a small window for async enqueue to complete
	time.Sleep(500 * time.Millisecond)

	pendingProvisionTasks, err := inspector.ListPendingTasks(queue.QueueLow)
	require.NoError(t, err, "should be able to list pending tasks in 'low' queue")

	foundProvision := false
	for _, task := range pendingProvisionTasks {
		if task.Type == queue.TaskTypeProvision {
			var payload queue.ProvisionJobPayload
			err := json.Unmarshal(task.Payload, &payload)
			require.NoError(t, err)
			if payload.WorkloadID == workloadID {
				foundProvision = true
				assert.Equal(t, "PROVISION", payload.Action)
				t.Logf("  ✓ Provision job found: WorkloadID=%s, Action=%s",
					payload.WorkloadID, payload.Action)
			}
		}
	}
	assert.True(t, foundProvision,
		"provision job for workload should exist in Redis 'low' queue")

	// ══════════════════════════════════════════════════════════════
	// STEP 7: Assert Build Job Exists in Redis Queue
	// ══════════════════════════════════════════════════════════════
	t.Log("Step 7: Verifying build job in Redis queue...")

	pendingBuildTasks, err := inspector.ListPendingTasks(queue.QueueDefault)
	require.NoError(t, err, "should be able to list pending tasks in 'default' queue")

	foundBuild := false
	for _, task := range pendingBuildTasks {
		if task.Type == queue.TaskTypeBuild {
			var payload queue.BuildJobPayload
			err := json.Unmarshal(task.Payload, &payload)
			require.NoError(t, err)
			if payload.WorkloadID == workloadID {
				foundBuild = true
				assert.Equal(t, "abcdef", payload.CommitSHA)
				t.Logf("  ✓ Build job found: WorkloadID=%s, CommitSHA=%s",
					payload.WorkloadID, payload.CommitSHA)
			}
		}
	}
	assert.True(t, foundBuild,
		"build job for workload should exist in Redis 'default' queue")

	// ══════════════════════════════════════════════════════════════
	// STEP 8: Verify GetWorkloadStatus via Connect-RPC
	// ══════════════════════════════════════════════════════════════
	t.Log("Step 8: Verifying GetWorkloadStatus via Connect-RPC...")

	statusResp, err := client.GetWorkloadStatus(ctx, connect.NewRequest(&v1.GetWorkloadStatusRequest{
		WorkloadName: "test-workload",
	}))
	require.NoError(t, err, "GetWorkloadStatus should succeed")
	assert.Equal(t, "test-workload", statusResp.Msg.GetWorkloadName())
	assert.Equal(t, v1.WorkloadStatus_WORKLOAD_STATUS_PENDING, statusResp.Msg.GetStatus())
	t.Logf("  ✓ GetWorkloadStatus verified: Name=%s, Status=%s",
		statusResp.Msg.GetWorkloadName(), statusResp.Msg.GetStatus())

	// ══════════════════════════════════════════════════════════════
	// STEP 9: Verify Duplicate Workload Rejection
	// ══════════════════════════════════════════════════════════════
	t.Log("Step 9: Verifying duplicate workload rejection...")

	_, dupErr := client.SubmitWorkload(ctx, connect.NewRequest(&v1.SubmitWorkloadRequest{
		Workload: &v1.WorkloadDescriptor{
			Name: "test-workload", // same name as step 1
			Source: &v1.SourceSpec{
				Type:    "git",
				RepoUrl: "https://github.com/test/repo2",
				Branch:  "main",
			},
			Regions: &v1.RegionConfig{
				Primary: &v1.RegionTarget{
					Provider: v1.CloudProvider_CLOUD_PROVIDER_AWS,
					Region:   v1.Region_REGION_AWS_AP_SOUTH_1,
				},
				Standbys: []*v1.RegionTarget{
					{Provider: v1.CloudProvider_CLOUD_PROVIDER_AWS, Region: v1.Region_REGION_AWS_US_EAST_1},
				},
			},
			Compute:  &v1.ResourceSpec{Cpu: "1", Memory: "2Gi", Replicas: 1},
			Database: &v1.DatabaseSpec{Engine: "postgres", Version: "16", StorageGb: 10},
			Failover: &v1.FailoverConfig{RtoSeconds: 30, RpoSeconds: 5},
		},
	}))
	require.Error(t, dupErr, "duplicate workload name should be rejected")
	connectErr := new(connect.Error)
	require.ErrorAs(t, dupErr, &connectErr)
	assert.Equal(t, connect.CodeAlreadyExists, connectErr.Code(),
		"duplicate workload should return AlreadyExists")
	t.Log("  ✓ Duplicate workload correctly rejected with AlreadyExists")

	// ══════════════════════════════════════════════════════════════
	// STEP 10: Verify Invalid Webhook Signature Rejection
	// ══════════════════════════════════════════════════════════════
	t.Log("Step 10: Verifying invalid webhook signature rejection...")

	badSigReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		srv.URL+"/webhooks/github", bytes.NewReader(payloadBytes))
	require.NoError(t, err)
	badSigReq.Header.Set("Content-Type", "application/json")
	badSigReq.Header.Set("X-GitHub-Event", "push")
	badSigReq.Header.Set("X-Hub-Signature-256", "sha256=baaaaadbaaaaadbaaaaadbaaaaadbaaaaadbaaaaadbaaaaadbaaaaadbaaaaad00")

	badSigResp, err := srv.Client().Do(badSigReq)
	require.NoError(t, err)
	defer badSigResp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, badSigResp.StatusCode,
		"invalid signature should return 401")
	t.Log("  ✓ Invalid signature correctly rejected with 401")

	t.Log("")
	t.Log("E2E INTEGRATION TEST PASSED — All 10 steps verified")
}

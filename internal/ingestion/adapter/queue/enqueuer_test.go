//go:build integration
// +build integration

package queue_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"github.com/raftweave/raftweave/internal/ingestion/adapter/queue"
	"github.com/raftweave/raftweave/internal/ingestion/domain"
)

func TestEnqueueBuildJob_Success(t *testing.T) {
	opts := setupRedis(t)
	logger, _ := zap.NewDevelopment()
	tracer := noop.NewTracerProvider().Tracer("test")

	enqueuer := queue.NewAsynqJobEnqueuer(opts, tracer, logger)

	event := &domain.WebhookEvent{
		WorkloadName: "test-workload",
		RepoURL:      "https://github.com/test",
		Branch:       "main",
		CommitSHA:    "xyz",
		CommitMsg:    "msg",
		TriggeredAt:  time.Now(),
	}

	err := enqueuer.EnqueueBuildJob(context.Background(), "w-123", event)
	require.NoError(t, err)

	// Verify task in Redis using Asynq Inspector
	inspector := asynq.NewInspector(opts)
	tasks, err := inspector.ListPendingTasks(queue.QueueDefault)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)

	var payload queue.BuildJobPayload
	err = json.Unmarshal(tasks[0].Payload, &payload)
	require.NoError(t, err)

	assert.Equal(t, "w-123", payload.WorkloadID)
	assert.Equal(t, "test-workload", payload.WorkloadName)
	assert.Equal(t, "xyz", payload.CommitSHA)
}

func TestEnqueueProvisionJob_Success(t *testing.T) {
	opts := setupRedis(t)
	logger, _ := zap.NewDevelopment()
	tracer := noop.NewTracerProvider().Tracer("test")

	enqueuer := queue.NewAsynqJobEnqueuer(opts, tracer, logger)

	err := enqueuer.EnqueueProvisionJob(context.Background(), "w-456")
	require.NoError(t, err)

	// Verify task in Redis using Asynq Inspector
	inspector := asynq.NewInspector(opts)
	tasks, err := inspector.ListPendingTasks(queue.QueueLow)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)

	var payload queue.ProvisionJobPayload
	err = json.Unmarshal(tasks[0].Payload, &payload)
	require.NoError(t, err)

	assert.Equal(t, "w-456", payload.WorkloadID)
	assert.Equal(t, "PROVISION", payload.Action)
}

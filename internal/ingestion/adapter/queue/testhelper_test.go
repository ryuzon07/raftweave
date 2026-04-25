//go:build integration
// +build integration

package queue_test

import (
	"context"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupRedis(t *testing.T) asynq.RedisClientOpt {
	ctx := context.Background()

	redisContainer, err := redis.RunContainer(ctx,
		testcontainers.WithImage("docker.io/redis:7-alpine"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections"),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = redisContainer.Terminate(ctx)
	})

	uri, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	opts, err := asynq.ParseRedisURI(uri)
	require.NoError(t, err)

	return opts.(asynq.RedisClientOpt)
}

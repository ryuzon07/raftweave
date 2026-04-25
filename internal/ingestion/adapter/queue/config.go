package queue

import (
	"time"

	"github.com/hibiken/asynq"
)

const (
	QueueCritical = "critical" // weight 6 — failover commands
	QueueDefault  = "default"  // weight 3 — build jobs
	QueueLow      = "low"      // weight 1 — provisioning
)

var BuildJobOptions = []asynq.Option{
	asynq.MaxRetry(3),
	asynq.Timeout(30 * time.Minute), // builds can take time
	asynq.Queue(QueueDefault),
	asynq.Retention(24 * time.Hour), // keep completed tasks 24h for audit
}

var ProvisionJobOptions = []asynq.Option{
	asynq.MaxRetry(5),
	asynq.Timeout(10 * time.Minute),
	asynq.Queue(QueueLow),
	asynq.Retention(48 * time.Hour),
}

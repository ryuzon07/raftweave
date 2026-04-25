package ingestion

import (
	"go.opentelemetry.io/otel/metric"
)

type IngestionMetrics struct {
	WorkloadsSubmitted metric.Int64Counter
	WebhooksProcessed  metric.Int64Counter
	CredentialsAdded   metric.Int64Counter
	WebhookDuration    metric.Float64Histogram
	DBQueryDuration    metric.Float64Histogram
	QueueEnqueueDuration metric.Float64Histogram
}

func NewIngestionMetrics(meter metric.Meter) *IngestionMetrics {
	ws, _ := meter.Int64Counter("raftweave.ingestion.workloads_submitted_total")
	wp, _ := meter.Int64Counter("raftweave.ingestion.webhooks_processed_total")
	ca, _ := meter.Int64Counter("raftweave.ingestion.credentials_added_total")
	wd, _ := meter.Float64Histogram("raftweave.ingestion.webhook_processing_duration")
	dbd, _ := meter.Float64Histogram("raftweave.ingestion.db_query_duration_seconds")
	qed, _ := meter.Float64Histogram("raftweave.ingestion.queue_enqueue_duration_seconds")
	return &IngestionMetrics{
		WorkloadsSubmitted:   ws,
		WebhooksProcessed:    wp,
		CredentialsAdded:     ca,
		WebhookDuration:      wd,
		DBQueryDuration:      dbd,
		QueueEnqueueDuration: qed,
	}
}

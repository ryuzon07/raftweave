package metrics

import (
	"go.opentelemetry.io/otel/metric"
)

var (
	BuildsTotal       metric.Int64Counter
	BuildDurationMs   metric.Float64Histogram
	QueueDepth        metric.Int64Gauge
	KanikoJobDuration metric.Float64Histogram
	RegistryPushBytes metric.Int64Counter
	LogLinesPublished metric.Int64Counter
)

// InitMetrics initializes the OpenTelemetry metrics for the build pipeline.
func InitMetrics(meter metric.Meter) error {
	var err error

	BuildsTotal, err = meter.Int64Counter(
		"raftweave_builds_total",
		metric.WithDescription("Total number of builds executed"),
	)
	if err != nil {
		return err
	}

	BuildDurationMs, err = meter.Float64Histogram(
		"raftweave_build_duration_ms",
		metric.WithDescription("Duration of builds in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	QueueDepth, err = meter.Int64Gauge(
		"raftweave_build_queue_depth",
		metric.WithDescription("Current depth of the build queue"),
	)
	if err != nil {
		return err
	}

	KanikoJobDuration, err = meter.Float64Histogram(
		"raftweave_kaniko_job_duration_s",
		metric.WithDescription("Duration of the underlying kaniko job in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	RegistryPushBytes, err = meter.Int64Counter(
		"raftweave_registry_push_bytes_total",
		metric.WithDescription("Total bytes pushed to the registry"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return err
	}

	LogLinesPublished, err = meter.Int64Counter(
		"raftweave_log_lines_published_total",
		metric.WithDescription("Total number of log lines published to the broadcaster"),
	)
	return err
}

// Package observability provides OpenTelemetry initialization and instrumentation.
package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Provider wraps OTel tracer and meter providers.
type Provider struct {
	tracer trace.Tracer
}

// NewProvider creates and registers the OTel tracer and meter providers.
func NewProvider(ctx context.Context, serviceName, endpoint string) (*Provider, error) {
	_ = ctx
	_ = endpoint

	// TODO: Initialize OTLP exporter and tracer provider.
	tracer := otel.Tracer(serviceName)
	return &Provider{tracer: tracer}, nil
}

// Tracer returns the application tracer.
func (p *Provider) Tracer() trace.Tracer {
	return p.tracer
}

// Shutdown gracefully shuts down all OTel providers.
func (p *Provider) Shutdown(ctx context.Context) error {
	_ = ctx
	return nil // stub
}

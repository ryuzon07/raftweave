package observability

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel"
)

// NewTracingInterceptor returns a Connect-RPC interceptor that creates OTel spans.
func NewTracingInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			tracer := otel.Tracer("connectrpc")
			ctx, span := tracer.Start(ctx, req.Spec().Procedure)
			defer span.End()

			return next(ctx, req)
		}
	}
}

// TracingMiddleware wraps an http.Handler with OTel HTTP tracing.
func TracingMiddleware(next http.Handler) http.Handler {
	// TODO: Wrap with otelhttp.NewHandler
	return next
}

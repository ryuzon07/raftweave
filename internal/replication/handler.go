package replication

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	raftweavev1 "github.com/raftweave/raftweave/internal/gen/raftweave/v1"
	"github.com/raftweave/raftweave/internal/gen/raftweave/v1/raftweavev1connect"
)

// Handler implements the Connect-RPC ReplicationService.
type Handler struct {
	raftweavev1connect.UnimplementedReplicationServiceHandler
	mgr    ReplicationManager
	tracer trace.Tracer
}

// NewHandler creates a new replication handler.
func NewHandler(mgr ReplicationManager) *Handler {
	return &Handler{
		mgr:    mgr,
		tracer: otel.Tracer("internal/replication"),
	}
}

func (h *Handler) GetReplicationStatus(ctx context.Context, req *connect.Request[raftweavev1.GetReplicationStatusRequest]) (*connect.Response[raftweavev1.GetReplicationStatusResponse], error) {
	ctx, span := h.tracer.Start(ctx, "replication.Handler.GetReplicationStatus")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (h *Handler) StreamReplicationMetrics(ctx context.Context, req *connect.Request[raftweavev1.StreamReplicationMetricsRequest], stream *connect.ServerStream[raftweavev1.StreamReplicationMetricsResponse]) error {
	ctx, span := h.tracer.Start(ctx, "replication.Handler.StreamReplicationMetrics")
	defer span.End()

	_ = req
	return connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (h *Handler) PromoteStandby(ctx context.Context, req *connect.Request[raftweavev1.PromoteStandbyRequest]) (*connect.Response[raftweavev1.PromoteStandbyResponse], error) {
	ctx, span := h.tracer.Start(ctx, "replication.Handler.PromoteStandby")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

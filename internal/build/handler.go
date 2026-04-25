package build

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	raftweavev1 "github.com/raftweave/raftweave/internal/gen/raftweave/v1"
	"github.com/raftweave/raftweave/internal/gen/raftweave/v1/raftweavev1connect"
)

var tracer = otel.Tracer("internal/build")

// Handler implements the Connect-RPC BuildService.
type Handler struct {
	raftweavev1connect.UnimplementedBuildServiceHandler
	svc    Service
	tracer trace.Tracer
}

// NewHandler creates a new build handler.
func NewHandler(svc Service) *Handler {
	return &Handler{
		svc:    svc,
		tracer: tracer,
	}
}

func (h *Handler) TriggerBuild(ctx context.Context, req *connect.Request[raftweavev1.TriggerBuildRequest]) (*connect.Response[raftweavev1.TriggerBuildResponse], error) {
	ctx, span := h.tracer.Start(ctx, "build.Handler.TriggerBuild")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (h *Handler) StreamBuildLogs(ctx context.Context, req *connect.Request[raftweavev1.StreamBuildLogsRequest], stream *connect.ServerStream[raftweavev1.StreamBuildLogsResponse]) error {
	ctx, span := h.tracer.Start(ctx, "build.Handler.StreamBuildLogs")
	defer span.End()

	_ = req
	return connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (h *Handler) GetBuildResult(ctx context.Context, req *connect.Request[raftweavev1.GetBuildResultRequest]) (*connect.Response[raftweavev1.GetBuildResultResponse], error) {
	ctx, span := h.tracer.Start(ctx, "build.Handler.GetBuildResult")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

package ingestion

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	raftweavev1 "github.com/raftweave/raftweave/internal/gen/raftweave/v1"
	"github.com/raftweave/raftweave/internal/gen/raftweave/v1/raftweavev1connect"
)

var tracer = otel.Tracer("internal/ingestion")

// Handler implements the Connect-RPC IngestionService.
type Handler struct {
	raftweavev1connect.UnimplementedIngestionServiceHandler
	svc    Service
	tracer trace.Tracer
}

// NewHandler creates a new ingestion handler.
func NewHandler(svc Service) *Handler {
	return &Handler{
		svc:    svc,
		tracer: tracer,
	}
}

func (h *Handler) SubmitWorkload(ctx context.Context, req *connect.Request[raftweavev1.SubmitWorkloadRequest]) (*connect.Response[raftweavev1.SubmitWorkloadResponse], error) {
	ctx, span := h.tracer.Start(ctx, "ingestion.Handler.SubmitWorkload")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (h *Handler) AddCloudCredentials(ctx context.Context, req *connect.Request[raftweavev1.AddCloudCredentialsRequest]) (*connect.Response[raftweavev1.AddCloudCredentialsResponse], error) {
	ctx, span := h.tracer.Start(ctx, "ingestion.Handler.AddCloudCredentials")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (h *Handler) GetWorkloadStatus(ctx context.Context, req *connect.Request[raftweavev1.GetWorkloadStatusRequest]) (*connect.Response[raftweavev1.GetWorkloadStatusResponse], error) {
	ctx, span := h.tracer.Start(ctx, "ingestion.Handler.GetWorkloadStatus")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (h *Handler) ListWorkloads(ctx context.Context, req *connect.Request[raftweavev1.ListWorkloadsRequest]) (*connect.Response[raftweavev1.ListWorkloadsResponse], error) {
	ctx, span := h.tracer.Start(ctx, "ingestion.Handler.ListWorkloads")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

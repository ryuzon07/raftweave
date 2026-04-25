package provisioner

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	raftweavev1 "github.com/raftweave/raftweave/internal/gen/raftweave/v1"
	"github.com/raftweave/raftweave/internal/gen/raftweave/v1/raftweavev1connect"
)

// Handler implements the Connect-RPC ProvisionerService.
type Handler struct {
	raftweavev1connect.UnimplementedProvisionerServiceHandler
	prov   CloudProvisioner
	tracer trace.Tracer
}

// NewHandler creates a new provisioner handler.
func NewHandler(prov CloudProvisioner) *Handler {
	return &Handler{
		prov:   prov,
		tracer: otel.Tracer("internal/provisioner"),
	}
}

func (h *Handler) ProvisionWorkload(ctx context.Context, req *connect.Request[raftweavev1.ProvisionWorkloadRequest]) (*connect.Response[raftweavev1.ProvisionWorkloadResponse], error) {
	ctx, span := h.tracer.Start(ctx, "provisioner.Handler.ProvisionWorkload")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (h *Handler) ExecuteFailover(ctx context.Context, req *connect.Request[raftweavev1.ExecuteFailoverRequest]) (*connect.Response[raftweavev1.ExecuteFailoverResponse], error) {
	ctx, span := h.tracer.Start(ctx, "provisioner.Handler.ExecuteFailover")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (h *Handler) GetResourceStatus(ctx context.Context, req *connect.Request[raftweavev1.GetResourceStatusRequest]) (*connect.Response[raftweavev1.GetResourceStatusResponse], error) {
	ctx, span := h.tracer.Start(ctx, "provisioner.Handler.GetResourceStatus")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

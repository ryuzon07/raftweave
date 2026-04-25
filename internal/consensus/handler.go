package consensus

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	raftweavev1 "github.com/raftweave/raftweave/internal/gen/raftweave/v1"
	"github.com/raftweave/raftweave/internal/gen/raftweave/v1/raftweavev1connect"
)

// Handler implements the Connect-RPC ConsensusService.
type Handler struct {
	raftweavev1connect.UnimplementedConsensusServiceHandler
	raft   *Raft
	tracer trace.Tracer
}

// NewHandler creates a new consensus handler.
func NewHandler(raft *Raft) *Handler {
	return &Handler{
		raft:   raft,
		tracer: otel.Tracer("internal/consensus"),
	}
}

func (h *Handler) GetClusterState(ctx context.Context, req *connect.Request[raftweavev1.GetClusterStateRequest]) (*connect.Response[raftweavev1.GetClusterStateResponse], error) {
	ctx, span := h.tracer.Start(ctx, "consensus.Handler.GetClusterState")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (h *Handler) StreamClusterState(ctx context.Context, req *connect.Request[raftweavev1.StreamClusterStateRequest], stream *connect.ServerStream[raftweavev1.StreamClusterStateResponse]) error {
	ctx, span := h.tracer.Start(ctx, "consensus.Handler.StreamClusterState")
	defer span.End()

	_ = req
	return connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (h *Handler) RequestVote(ctx context.Context, req *connect.Request[raftweavev1.RequestVoteRequest]) (*connect.Response[raftweavev1.RequestVoteResponse], error) {
	ctx, span := h.tracer.Start(ctx, "consensus.Handler.RequestVote")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (h *Handler) AppendEntries(ctx context.Context, req *connect.Request[raftweavev1.AppendEntriesRequest]) (*connect.Response[raftweavev1.AppendEntriesResponse], error) {
	ctx, span := h.tracer.Start(ctx, "consensus.Handler.AppendEntries")
	defer span.End()

	_ = req
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

package handler

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/raftweave/raftweave/internal/ingestion/domain"
)

// mapDomainError converts domain-specific errors into corresponding Connect-RPC errors
func mapDomainError(err error) error {
	if err == nil {
		return nil
	}

	// Validation errors -> InvalidArgument
	var valErr domain.ValidationErrors
	if errors.As(err, &valErr) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	if errors.Is(err, domain.ErrInvalidWorkloadName) ||
		errors.Is(err, domain.ErrInvalidSource) ||
		errors.Is(err, domain.ErrInvalidRTO) ||
		errors.Is(err, domain.ErrInvalidRPO) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Not Found -> NotFound
	if errors.Is(err, domain.ErrWorkloadNotFound) {
		return connect.NewError(connect.CodeNotFound, err)
	}

	// Conflict -> AlreadyExists
	if errors.Is(err, domain.ErrWorkloadAlreadyExists) {
		return connect.NewError(connect.CodeAlreadyExists, err)
	}

	// Fallback to Internal
	return connect.NewError(connect.CodeInternal, err)
}
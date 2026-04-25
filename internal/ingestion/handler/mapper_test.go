package handler

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"

	"github.com/raftweave/raftweave/internal/ingestion/domain"
)

func TestMapDomainError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr connect.Code
	}{
		{
			name:    "nil error",
			err:     nil,
			wantErr: connect.CodeUnknown, // when nil, return is nil, won't check Code
		},
		{
			name:    "ValidationErrors -> CodeInvalidArgument",
			err:     domain.ValidationErrors{{Field: "Name", Message: "bad"}},
			wantErr: connect.CodeInvalidArgument,
		},
		{
			name:    "ErrInvalidWorkloadName -> CodeInvalidArgument",
			err:     domain.ErrInvalidWorkloadName,
			wantErr: connect.CodeInvalidArgument,
		},
		{
			name:    "ErrWorkloadNotFound -> CodeNotFound",
			err:     domain.ErrWorkloadNotFound,
			wantErr: connect.CodeNotFound,
		},
		{
			name:    "ErrWorkloadAlreadyExists -> CodeAlreadyExists",
			err:     domain.ErrWorkloadAlreadyExists,
			wantErr: connect.CodeAlreadyExists,
		},
		{
			name:    "Generic wrapped error -> CodeInternal",
			err:     errors.New("some random db failure"),
			wantErr: connect.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := mapDomainError(tt.err)
			if tt.err == nil {
				assert.Nil(t, res)
				return
			}
			assert.NotNil(t, res)
			
			var connectErr *connect.Error
			requireErr := errors.As(res, &connectErr)
			assert.True(t, requireErr, "Expected connect.Error type")
			if requireErr {
				assert.Equal(t, tt.wantErr, connectErr.Code())
			}
		})
	}
}
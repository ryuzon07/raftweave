package auth

import (
	"context"
	"errors"

	"connectrpc.com/connect"
)

// NewAuthInterceptor returns a Connect-RPC interceptor that validates session tokens.
func NewAuthInterceptor(authSvc AuthService) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			_ = authSvc
			// TODO: Extract Bearer token from Authorization header, validate session.
			return nil, connect.NewError(connect.CodeUnimplemented, errors.New("auth not implemented"))
		}
	}
}

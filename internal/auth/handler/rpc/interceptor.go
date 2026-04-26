package rpc

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	jwtadapter "github.com/raftweave/raftweave/internal/auth/adapter/jwt"
)

// NewAuthInterceptor creates a Connect interceptor that extracts JWT from cookies.
func NewAuthInterceptor(issuer jwtadapter.Issuer) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Skip auth for RequestOTP and VerifyOTP
			procedure := req.Spec().Procedure
			if procedure == "/auth.v1.AuthService/RequestOTP" || procedure == "/auth.v1.AuthService/VerifyOTP" {
				return next(ctx, req)
			}

			// Try to get token from Authorization header or Cookie
			tokenStr := ""
			if auth := req.Header().Get("Authorization"); auth != "" {
				if len(auth) > 7 && auth[:7] == "Bearer " {
					tokenStr = auth[7:]
				}
			}

			if tokenStr == "" {
				// Accessing cookies in Connect interceptors requires the underlying http.Request.
				// However, Connect abstracts this away. We rely on the CORS middleware 
				// to have allowed the cookie, and we check the cookie header manually.
				cookieHeader := req.Header().Get("Cookie")
				if cookieHeader != "" {
					header := http.Header{"Cookie": {cookieHeader}}
					request := &http.Request{Header: header}
					if cookie, err := request.Cookie("raftweave_at"); err == nil {
						tokenStr = cookie.Value
					}
				}
			}

			if tokenStr == "" {
				return next(ctx, req)
			}

			claims, err := issuer.Validate(ctx, tokenStr)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
			}

			// Inject claims into context
			ctx = context.WithValue(ctx, ContextKeyClaims, claims)
			return next(ctx, req)
		}
	}
}

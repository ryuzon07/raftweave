// Package middleware provides Connect-RPC interceptors and HTTP middleware
// for JWT authentication, RBAC enforcement, rate limiting, and security headers.
package middleware

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"

	jwtadapter "github.com/raftweave/raftweave/internal/auth/adapter/jwt"
	"github.com/raftweave/raftweave/internal/auth/domain"
)

// contextKey type prevents collisions with other context values.
type contextKey string

const (
	// ContextKeyUserID stores the authenticated user ID.
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyClaims stores the validated JWT claims.
	ContextKeyClaims contextKey = "jwt_claims"
	// ContextKeySessionID stores the session ID.
	ContextKeySessionID contextKey = "session_id"
)

// PublicRoutes lists endpoints that do not require authentication.
var PublicRoutes = []string{
	"/auth.v1.AuthService/RequestOTP",
	"/auth.v1.AuthService/VerifyOTP",
	"/auth.v1.AuthService/RefreshToken",
}

// NewAuthInterceptor returns a Connect-RPC interceptor that:
// 1. Extracts JWT from Authorization: Bearer <token> header OR raftweave_at cookie
// 2. Validates the JWT using the Issuer's public key
// 3. Injects Claims into the request context
// 4. Returns Unauthenticated if token is missing/invalid/expired
func NewAuthInterceptor(issuer jwtadapter.Issuer, publicRoutes []string) connect.UnaryInterceptorFunc {
	publicSet := make(map[string]struct{}, len(publicRoutes))
	for _, r := range publicRoutes {
		publicSet[r] = struct{}{}
	}

	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			procedure := req.Spec().Procedure

			// Allow public routes without authentication.
			if _, ok := publicSet[procedure]; ok {
				return next(ctx, req)
			}

			// Extract token: prefer Authorization header over cookie.
			tokenStr := extractToken(req)
			if tokenStr == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing authentication token"))
			}

			// Validate JWT.
			claims, err := issuer.Validate(ctx, tokenStr)
			if err != nil {
				if errors.Is(err, domain.ErrTokenExpired) {
					return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("token expired"))
				}
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
			}

			// Inject claims into context.
			ctx = context.WithValue(ctx, ContextKeyClaims, claims)
			ctx = context.WithValue(ctx, ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeySessionID, claims.SessionID)

			return next(ctx, req)
		}
	}
}

// NewRequireRole returns an interceptor that checks the user has at least
// the given role in the workspace specified by the x-workspace-id header.
func NewRequireRole(minRole domain.Role) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			claims := ClaimsFromContext(ctx)
			if claims == nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("not authenticated"))
			}

			workspaceID := req.Header().Get("X-Workspace-Id")
			if workspaceID == "" {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("x-workspace-id header required"))
			}

			if err := RequireWorkspaceRole(ctx, workspaceID, minRole); err != nil {
				return nil, connect.NewError(connect.CodePermissionDenied, errors.New("insufficient permissions"))
			}

			return next(ctx, req)
		}
	}
}

// ClaimsFromContext retrieves the validated JWT claims from the context.
func ClaimsFromContext(ctx context.Context) *jwtadapter.Claims {
	c, _ := ctx.Value(ContextKeyClaims).(*jwtadapter.Claims)
	return c
}

// RequireWorkspaceRole checks that the user has at least the given role
// in the specified workspace. Returns ErrInsufficientRole if not.
func RequireWorkspaceRole(ctx context.Context, workspaceID string, minRole domain.Role) error {
	claims := ClaimsFromContext(ctx)
	if claims == nil {
		return domain.ErrInsufficientRole
	}

	userRoleStr, ok := claims.Roles[workspaceID]
	if !ok {
		return domain.ErrInsufficientRole
	}

	userRole := domain.Role(userRoleStr)
	if !isRoleSufficient(userRole, minRole) {
		return domain.ErrInsufficientRole
	}

	return nil
}

// extractToken extracts the JWT from the request.
// Priority: 1. Authorization: Bearer <token>, 2. raftweave_at cookie.
// Never accepts tokens from query parameters.
func extractToken(req connect.AnyRequest) string {
	// 1. Authorization header.
	auth := req.Header().Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// 2. Cookie fallback.
	cookies := req.Header().Values("Cookie")
	for _, c := range cookies {
		for _, part := range strings.Split(c, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "raftweave_at=") {
				return strings.TrimPrefix(part, "raftweave_at=")
			}
		}
	}

	return ""
}

// roleHierarchy defines the permission level of each role.
// Higher number = more permissions.
var roleHierarchy = map[domain.Role]int{
	domain.RoleViewer: 1,
	domain.RoleMember: 2,
	domain.RoleAdmin:  3,
	domain.RoleOwner:  4,
}

// isRoleSufficient checks if userRole meets or exceeds minRole.
func isRoleSufficient(userRole, minRole domain.Role) bool {
	userLevel, ok1 := roleHierarchy[userRole]
	minLevel, ok2 := roleHierarchy[minRole]
	if !ok1 || !ok2 {
		return false
	}
	return userLevel >= minLevel
}

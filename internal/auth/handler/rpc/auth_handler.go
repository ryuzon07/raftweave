// Package rpc implements Connect-RPC handlers for the AuthService.
package rpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	jwtadapter "github.com/raftweave/raftweave/internal/auth/adapter/jwt"
	githubprovider "github.com/raftweave/raftweave/internal/auth/adapter/oauth/github"
	"github.com/raftweave/raftweave/internal/auth/adapter/otp"
	redisadapter "github.com/raftweave/raftweave/internal/auth/adapter/redis"
	"github.com/raftweave/raftweave/internal/auth/domain"
	authv1 "github.com/raftweave/raftweave/internal/gen/auth/v1"
	"go.uber.org/zap"
)

// contextKey type prevents collisions with other context values.
type contextKey string

const (
	// ContextKeyClaims is the context key for JWT claims.
	ContextKeyClaims contextKey = "jwt_claims"
)

// ClaimsFromContext retrieves validated JWT claims from context.
func ClaimsFromContext(ctx context.Context) *jwtadapter.Claims {
	c, _ := ctx.Value(ContextKeyClaims).(*jwtadapter.Claims)
	return c
}

// AuthHandler implements the AuthService Connect-RPC service.
type AuthHandler struct {
	otpGen      otp.Generator
	jwtIssuer   jwtadapter.Issuer
	tokenStore  redisadapter.RefreshTokenStore
	userRepo    domain.UserRepository
	memberRepo  domain.MembershipRepository
	sessionRepo domain.SessionRepository
	githubProv  githubprovider.Provider
	mailer      otp.Mailer
	logger      *zap.Logger
}

// NewAuthHandler creates a new Connect-RPC auth handler.
func NewAuthHandler(
	og otp.Generator, jwt jwtadapter.Issuer, ts redisadapter.RefreshTokenStore,
	ur domain.UserRepository, mr domain.MembershipRepository,
	sr domain.SessionRepository, gh githubprovider.Provider, m otp.Mailer,
	l *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		otpGen: og, jwtIssuer: jwt, tokenStore: ts,
		userRepo: ur, memberRepo: mr, sessionRepo: sr, githubProv: gh, mailer: m,
		logger: l,
	}
}

// RequestOTP initiates an OTP challenge.
func (h *AuthHandler) RequestOTP(ctx context.Context, req *connect.Request[authv1.RequestOTPRequest]) (*connect.Response[authv1.RequestOTPResponse], error) {
	cid, err := h.otpGen.Issue(ctx, req.Msg.GetEmail(), domain.OTPPurposeLogin)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("authentication service error"))
	}
	return connect.NewResponse(&authv1.RequestOTPResponse{ChallengeId: cid, ExpiresIn: 600}), nil
}

// VerifyOTP verifies the OTP and issues tokens.
func (h *AuthHandler) VerifyOTP(ctx context.Context, req *connect.Request[authv1.VerifyOTPRequest]) (*connect.Response[authv1.AuthTokenResponse], error) {
	email, err := h.otpGen.Verify(ctx, req.Msg.GetChallengeId(), req.Msg.GetCode())
	if err != nil {
		return nil, mapDomainError(err)
	}
	user, err := h.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			user = &domain.User{
				ID: generateID(), Email: email, Name: email,
				Provider: domain.ProviderEmail, ProviderID: email,
				IsEmailVerified: true, IsActive: true,
			}
			if cErr := h.userRepo.Create(ctx, user); cErr != nil {
				return nil, connect.NewError(connect.CodeInternal, errors.New("authentication service error"))
			}
		} else {
			return nil, connect.NewError(connect.CodeInternal, errors.New("authentication service error"))
		}
	}
	_ = h.userRepo.UpdateLastLogin(ctx, user.ID)
	return h.issueTokenResponse(ctx, req.Header(), user)
}

// RefreshToken rotates the refresh token and issues a new access token.
func (h *AuthHandler) RefreshToken(ctx context.Context, req *connect.Request[authv1.RefreshTokenRequest]) (*connect.Response[authv1.AuthTokenResponse], error) {
	ip := extractIP(req.Header())
	fp := redisadapter.BuildFingerprint(req.Header().Get("User-Agent"), ip)

	newRT, newSID, err := h.tokenStore.Rotate(ctx, req.Msg.GetRefreshToken(), fp)
	if err != nil {
		if errors.Is(err, domain.ErrSessionFingerprintMismatch) && h.mailer != nil {
			go func() { _ = h.mailer.SendSecurityAlert(context.Background(), "", "Suspicious login — all sessions revoked") }()
		}
		return nil, mapDomainError(err)
	}

	claims := ClaimsFromContext(ctx)
	if claims == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("session invalid"))
	}
	user, err := h.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("authentication service error"))
	}

	roles := h.buildRoles(ctx, user.ID)
	at, err := h.jwtIssuer.IssueAccessToken(ctx, user, newSID, roles)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("authentication service error"))
	}

	return connect.NewResponse(&authv1.AuthTokenResponse{
		AccessToken: at, RefreshToken: newRT,
		AccessTokenExpiresIn: 900, RefreshTokenExpiresIn: 604800,
		User: userToProto(user),
	}), nil
}

// Logout revokes a single refresh token.
func (h *AuthHandler) Logout(ctx context.Context, req *connect.Request[authv1.LogoutRequest]) (*connect.Response[authv1.LogoutResponse], error) {
	_ = h.tokenStore.Revoke(ctx, req.Msg.GetRefreshToken())
	return connect.NewResponse(&authv1.LogoutResponse{Success: true}), nil
}

// LogoutAll revokes all sessions for the authenticated user.
func (h *AuthHandler) LogoutAll(ctx context.Context, _ *connect.Request[authv1.LogoutAllRequest]) (*connect.Response[authv1.LogoutResponse], error) {
	claims := ClaimsFromContext(ctx)
	if claims == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("not authenticated"))
	}
	_ = h.tokenStore.RevokeAll(ctx, claims.UserID)
	return connect.NewResponse(&authv1.LogoutResponse{Success: true}), nil
}

// GetMe returns the authenticated user's profile.
func (h *AuthHandler) GetMe(ctx context.Context, _ *connect.Request[authv1.GetMeRequest]) (*connect.Response[authv1.GetMeResponse], error) {
	claims := ClaimsFromContext(ctx)
	if claims == nil {
		h.logger.Warn("GetMe called without valid claims")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("not authenticated"))
	}
	
	h.logger.Info("GetMe successful", zap.String("user_id", claims.UserID))
	
	user, err := h.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		h.logger.Error("GetMe failed to get user", zap.String("user_id", claims.UserID), zap.Error(err))
		return nil, mapDomainError(err)
	}
	memberships, _ := h.memberRepo.GetUserWorkspaces(ctx, user.ID)
	var roles []*authv1.WorkspaceRole
	for _, m := range memberships {
		roles = append(roles, &authv1.WorkspaceRole{WorkspaceId: m.WorkspaceID, Role: string(m.Role)})
	}
	return connect.NewResponse(&authv1.GetMeResponse{User: userToProto(user), Roles: roles}), nil
}

// ListSessions returns active sessions for the authenticated user.
func (h *AuthHandler) ListSessions(ctx context.Context, _ *connect.Request[authv1.ListSessionsRequest]) (*connect.Response[authv1.ListSessionsResponse], error) {
	claims := ClaimsFromContext(ctx)
	if claims == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("not authenticated"))
	}
	sessions, err := h.sessionRepo.GetByUserID(ctx, claims.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("authentication service error"))
	}
	var infos []*authv1.SessionInfo
	for _, s := range sessions {
		if s.IsRevoked { continue }
		fp := s.Fingerprint
		if len(fp) > 8 { fp = fp[:8] } // Truncate for privacy.
		infos = append(infos, &authv1.SessionInfo{
			SessionId: s.ID, Fingerprint: fp,
			CreatedAt: timestamppb.New(s.CreatedAt), ExpiresAt: timestamppb.New(s.ExpiresAt),
			IsCurrent: s.ID == claims.SessionID,
		})
	}
	return connect.NewResponse(&authv1.ListSessionsResponse{Sessions: infos}), nil
}

// ListUserRepos lists GitHub repos for the authenticated user.
func (h *AuthHandler) ListUserRepos(ctx context.Context, _ *connect.Request[authv1.ListUserReposRequest]) (*connect.Response[authv1.ListUserReposResponse], error) {
	claims := ClaimsFromContext(ctx)
	if claims == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("not authenticated"))
	}
	repos, err := h.githubProv.ListUserRepos(ctx, claims.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to list repositories"))
	}
	var out []*authv1.GitHubRepository
	for _, r := range repos {
		out = append(out, &authv1.GitHubRepository{
			Id: r.ID, FullName: r.FullName, HtmlUrl: r.HTMLURL,
			DefaultBranch: r.DefaultBranch, IsPrivate: r.Private, CloneUrl: r.CloneURL,
		})
	}
	return connect.NewResponse(&authv1.ListUserReposResponse{Repos: out}), nil
}

// RevokeRepoAccess is a stub for future implementation.
func (h *AuthHandler) RevokeRepoAccess(_ context.Context, _ *connect.Request[authv1.RevokeRepoAccessRequest]) (*connect.Response[authv1.RevokeRepoAccessResponse], error) {
	return connect.NewResponse(&authv1.RevokeRepoAccessResponse{Revoked: true}), nil
}

// --- Helpers ---

func (h *AuthHandler) issueTokenResponse(ctx context.Context, hdrs http.Header, user *domain.User) (*connect.Response[authv1.AuthTokenResponse], error) {
	ip := hdrs.Get("X-Forwarded-For")
	if ip == "" { ip = hdrs.Get("X-Real-Ip") }
	fp := redisadapter.BuildFingerprint(hdrs.Get("User-Agent"), ip)
	roles := h.buildRoles(ctx, user.ID)
	sid := generateID()

	rt, err := h.tokenStore.Issue(ctx, sid, user.ID, fp)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("authentication service error"))
	}
	at, err := h.jwtIssuer.IssueAccessToken(ctx, user, sid, roles)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("authentication service error"))
	}
	return connect.NewResponse(&authv1.AuthTokenResponse{
		AccessToken: at, RefreshToken: rt,
		AccessTokenExpiresIn: 900, RefreshTokenExpiresIn: 604800,
		User: userToProto(user),
	}), nil
}

func (h *AuthHandler) buildRoles(ctx context.Context, userID string) map[string]string {
	roles := make(map[string]string)
	if h.memberRepo != nil {
		ms, _ := h.memberRepo.GetUserWorkspaces(ctx, userID)
		for _, m := range ms { roles[m.WorkspaceID] = string(m.Role) }
	}
	return roles
}

func userToProto(u *domain.User) *authv1.UserInfo {
	info := &authv1.UserInfo{
		UserId: u.ID, Email: u.Email, Name: u.Name,
		AvatarUrl: u.AvatarURL, Provider: string(u.Provider),
		EmailVerified: u.IsEmailVerified, CreatedAt: timestamppb.New(u.CreatedAt),
	}
	if u.LastLoginAt != nil { info.LastLoginAt = timestamppb.New(*u.LastLoginAt) }
	return info
}

func mapDomainError(err error) *connect.Error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	case errors.Is(err, domain.ErrTokenExpired), errors.Is(err, domain.ErrTokenInvalid):
		return connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
	case errors.Is(err, domain.ErrSessionFingerprintMismatch), errors.Is(err, domain.ErrSessionNotFound),
		errors.Is(err, domain.ErrSessionExpired), errors.Is(err, domain.ErrSessionRevoked):
		return connect.NewError(connect.CodeUnauthenticated, errors.New("session invalid"))
	case errors.Is(err, domain.ErrOTPInvalid):
		return connect.NewError(connect.CodeInvalidArgument, errors.New("invalid code"))
	case errors.Is(err, domain.ErrOTPExpired):
		return connect.NewError(connect.CodeDeadlineExceeded, errors.New("code expired"))
	case errors.Is(err, domain.ErrOTPMaxAttemptsReached):
		return connect.NewError(connect.CodeResourceExhausted, errors.New("max attempts reached"))
	case errors.Is(err, domain.ErrOTPAlreadyUsed):
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("code already used"))
	case errors.Is(err, domain.ErrOTPNotFound):
		return connect.NewError(connect.CodeNotFound, errors.New("challenge not found"))
	case errors.Is(err, domain.ErrInsufficientRole):
		return connect.NewError(connect.CodePermissionDenied, errors.New("insufficient permissions"))
	default:
		return connect.NewError(connect.CodeInternal, errors.New("authentication service error"))
	}
}

func extractIP(h interface{ Get(string) string }) string {
	if v := h.Get("X-Forwarded-For"); v != "" { return v }
	if v := h.Get("X-Real-Ip"); v != "" { return v }
	return "unknown"
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

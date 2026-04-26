// Package http implements OAuth callback HTTP handlers.
// OAuth callbacks MUST be HTTP handlers (not Connect-RPC) because browsers
// need to follow redirects and receive cookies.
package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	githubprovider "github.com/raftweave/raftweave/internal/auth/adapter/oauth/github"
	googleprovider "github.com/raftweave/raftweave/internal/auth/adapter/oauth/google"
	redisadapter "github.com/raftweave/raftweave/internal/auth/adapter/redis"
	jwtadapter "github.com/raftweave/raftweave/internal/auth/adapter/jwt"
	"github.com/raftweave/raftweave/internal/auth/domain"
	"go.uber.org/zap"
)

// OAuthHandler serves OAuth login and callback endpoints.
type OAuthHandler struct {
	github       githubprovider.Provider
	google       googleprovider.Provider
	jwtIssuer    jwtadapter.Issuer
	tokenStore   redisadapter.RefreshTokenStore
	userRepo     domain.UserRepository
	memberRepo   domain.MembershipRepository
	cookieDomain string
	dashboardURL string
	logger       *zap.Logger
}

// NewOAuthHandler creates a new OAuth HTTP handler.
func NewOAuthHandler(
	gh githubprovider.Provider,
	gg googleprovider.Provider,
	jwt jwtadapter.Issuer,
	ts redisadapter.RefreshTokenStore,
	ur domain.UserRepository,
	mr domain.MembershipRepository,
	cookieDomain, dashboardURL string,
	logger *zap.Logger,
) *OAuthHandler {
	return &OAuthHandler{
		github:       gh,
		google:       gg,
		jwtIssuer:    jwt,
		tokenStore:   ts,
		userRepo:     ur,
		memberRepo:   mr,
		cookieDomain: cookieDomain,
		dashboardURL: dashboardURL,
		logger:       logger,
	}
}

// RegisterRoutes registers OAuth routes on the given mux.
func (h *OAuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/github/login", h.GitHubLogin)
	mux.HandleFunc("GET /auth/github/callback", h.GitHubCallback)
	mux.HandleFunc("GET /auth/google/login", h.GoogleLogin)
	mux.HandleFunc("GET /auth/google/callback", h.GoogleCallback)
}

// GitHubLogin redirects to GitHub OAuth URL.
func (h *OAuthHandler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	redirectURI := h.getRedirectURI(r, "github")
	url, _, err := h.github.AuthURL(r.Context(), redirectURI)
	if err != nil {
		http.Error(w, "Failed to initiate GitHub login", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GitHubCallback handles the GitHub OAuth callback.
func (h *OAuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "Missing code or state", http.StatusBadRequest)
		return
	}

	redirectURI := h.getRedirectURI(r, "github")
	
	// Use a 30s timeout for the GitHub exchange and user fetch
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	user, err := h.github.HandleCallback(ctx, code, state, redirectURI)
	if err != nil {
		h.logger.Error("GitHub authentication failed", zap.Error(err))
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	h.logger.Info("GitHub authentication successful", zap.String("email", user.Email))
	h.issueTokensAndRedirect(w, r, user)
}

// GoogleLogin redirects to Google OAuth URL.
func (h *OAuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	redirectURI := h.getRedirectURI(r, "google")
	url, _, err := h.google.AuthURL(r.Context(), redirectURI)
	if err != nil {
		http.Error(w, "Failed to initiate Google login", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the Google OAuth callback.
func (h *OAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "Missing code or state", http.StatusBadRequest)
		return
	}

	redirectURI := h.getRedirectURI(r, "google")
	user, err := h.google.HandleCallback(r.Context(), code, state, redirectURI)
	if err != nil {
		h.logger.Error("Google authentication failed", zap.Error(err))
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	h.issueTokensAndRedirect(w, r, user)
}

func (h *OAuthHandler) getRedirectURI(r *http.Request, provider string) string {
	scheme := "https://"
	host := r.Host
	if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
		scheme = "http://"
	}
	
	// Default to standard auth path. 
	// We don't include /api here because the Ingress is configured to handle /auth directly,
	// and the OAuth provider needs a stable callback URL.
	return scheme + host + "/auth/" + provider + "/callback"
}

// issueTokensAndRedirect creates JWT + refresh token, sets cookies, and redirects.
func (h *OAuthHandler) issueTokensAndRedirect(w http.ResponseWriter, r *http.Request, user *domain.User) {
	ctx := r.Context()

	// Build roles map from memberships.
	roles := make(map[string]string)
	if h.memberRepo != nil {
		memberships, err := h.memberRepo.GetUserWorkspaces(ctx, user.ID)
		if err == nil {
			for _, m := range memberships {
				roles[m.WorkspaceID] = string(m.Role)
			}
		}
	}

	fingerprint := redisadapter.BuildFingerprint(r.UserAgent(), extractClientIP(r))
	sessionID := user.ID + "-session"

	h.logger.Info("Issuing tokens for user", zap.String("user_id", user.ID), zap.String("session_id", sessionID))

	// Issue refresh token.
	refreshToken, err := h.tokenStore.Issue(ctx, sessionID, user.ID, fingerprint)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Issue access token.
	accessToken, err := h.jwtIssuer.IssueAccessToken(ctx, user, sessionID, roles)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set secure cookies.
	secure := true
	if strings.HasPrefix(r.Host, "localhost") || strings.HasPrefix(r.Host, "127.0.0.1") {
		secure = false
	}

	http.SetCookie(w, &http.Cookie{
		Name: "raftweave_at", Value: accessToken,
		Path: "/", Domain: h.cookieDomain,
		MaxAge: 900, Secure: secure, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, // Lax is better for OAuth redirects
	})
	http.SetCookie(w, &http.Cookie{
		Name: "raftweave_rt", Value: refreshToken,
		Path: "/auth/token", Domain: h.cookieDomain,
		MaxAge: 604800, Secure: secure, HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, h.dashboardURL+"?login=success", http.StatusTemporaryRedirect)
}

// extractClientIP extracts the client IP, checking X-Forwarded-For first.
func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

// mapDomainError converts domain errors to appropriate HTTP status codes.
func mapDomainError(err error) int {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrTokenExpired), errors.Is(err, domain.ErrTokenInvalid),
		errors.Is(err, domain.ErrSessionFingerprintMismatch):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrInsufficientRole):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

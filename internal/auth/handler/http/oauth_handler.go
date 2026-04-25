// Package http implements OAuth callback HTTP handlers.
// OAuth callbacks MUST be HTTP handlers (not Connect-RPC) because browsers
// need to follow redirects and receive cookies.
package http

import (
	"errors"
	"net/http"

	githubprovider "github.com/raftweave/raftweave/internal/auth/adapter/oauth/github"
	googleprovider "github.com/raftweave/raftweave/internal/auth/adapter/oauth/google"
	redisadapter "github.com/raftweave/raftweave/internal/auth/adapter/redis"
	jwtadapter "github.com/raftweave/raftweave/internal/auth/adapter/jwt"
	"github.com/raftweave/raftweave/internal/auth/domain"
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
) *OAuthHandler {
	return &OAuthHandler{
		github: gh, google: gg, jwtIssuer: jwt, tokenStore: ts,
		userRepo: ur, memberRepo: mr,
		cookieDomain: cookieDomain, dashboardURL: dashboardURL,
	}
}

// RegisterRoutes registers OAuth routes on the given mux.
func (h *OAuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/login/github", h.GitHubLogin)
	mux.HandleFunc("GET /auth/callback/github", h.GitHubCallback)
	mux.HandleFunc("GET /auth/login/google", h.GoogleLogin)
	mux.HandleFunc("GET /auth/callback/google", h.GoogleCallback)
}

// GitHubLogin redirects to GitHub OAuth URL.
func (h *OAuthHandler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	redirectURI := "https://" + r.Host + "/auth/callback/github"
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

	redirectURI := "https://" + r.Host + "/auth/callback/github"
	user, err := h.github.HandleCallback(r.Context(), code, state, redirectURI)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	h.issueTokensAndRedirect(w, r, user)
}

// GoogleLogin redirects to Google OAuth URL.
func (h *OAuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	redirectURI := "https://" + r.Host + "/auth/callback/google"
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

	redirectURI := "https://" + r.Host + "/auth/callback/google"
	user, err := h.google.HandleCallback(r.Context(), code, state, redirectURI)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	h.issueTokensAndRedirect(w, r, user)
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
	http.SetCookie(w, &http.Cookie{
		Name: "raftweave_at", Value: accessToken,
		Path: "/", Domain: h.cookieDomain,
		MaxAge: 900, Secure: true, HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name: "raftweave_rt", Value: refreshToken,
		Path: "/auth/token", Domain: h.cookieDomain,
		MaxAge: 604800, Secure: true, HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
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

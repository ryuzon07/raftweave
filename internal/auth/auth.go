// Package auth provides OAuth2 authentication via GitHub and Google.
package auth

import "context"

// AuthService defines authentication and session management.
type AuthService interface {
	// AuthenticateGitHub handles the GitHub OAuth2 callback.
	AuthenticateGitHub(ctx context.Context, code string) (*Session, error)
	// AuthenticateGoogle handles the Google OAuth2 callback.
	AuthenticateGoogle(ctx context.Context, code string) (*Session, error)
	// ValidateSession validates an existing session token.
	ValidateSession(ctx context.Context, token string) (*User, error)
	// RevokeSession invalidates a session.
	RevokeSession(ctx context.Context, token string) error
}

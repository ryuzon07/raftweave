package auth

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleAuth handles Google OAuth2 authentication.
type GoogleAuth struct {
	config *oauth2.Config
}

// NewGoogleAuth creates a new Google OAuth2 handler.
func NewGoogleAuth(clientID, clientSecret, redirectURL string) *GoogleAuth {
	return &GoogleAuth{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     google.Endpoint,
			Scopes:       []string{"openid", "email", "profile"},
		},
	}
}

// AuthURL returns the URL to redirect the user for Google login.
func (g *GoogleAuth) AuthURL(state string) string {
	return g.config.AuthCodeURL(state)
}

// Exchange exchanges an authorization code for a token.
func (g *GoogleAuth) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return g.config.Exchange(ctx, code)
}

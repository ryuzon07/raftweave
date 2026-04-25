package auth

import (
	"context"

	"golang.org/x/oauth2"
)

// GitHubAuth handles GitHub OAuth2 authentication.
type GitHubAuth struct {
	config *oauth2.Config
}

// NewGitHubAuth creates a new GitHub OAuth2 handler.
func NewGitHubAuth(clientID, clientSecret, redirectURL string) *GitHubAuth {
	return &GitHubAuth{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
			Scopes: []string{"user:email"},
		},
	}
}

// AuthURL returns the URL to redirect the user for GitHub login.
func (g *GitHubAuth) AuthURL(state string) string {
	return g.config.AuthCodeURL(state)
}

// Exchange exchanges an authorization code for a token.
func (g *GitHubAuth) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return g.config.Exchange(ctx, code)
}

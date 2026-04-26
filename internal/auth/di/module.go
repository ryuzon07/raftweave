package di

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"connectrpc.com/connect"
	"github.com/raftweave/raftweave/internal/auth/adapter/crypto"
	jwtadapter "github.com/raftweave/raftweave/internal/auth/adapter/jwt"
	"github.com/raftweave/raftweave/internal/auth/adapter/oauth"
	githubprovider "github.com/raftweave/raftweave/internal/auth/adapter/oauth/github"
	googleprovider "github.com/raftweave/raftweave/internal/auth/adapter/oauth/google"
	postgresadapter "github.com/raftweave/raftweave/internal/auth/adapter/postgres"
	redisadapter "github.com/raftweave/raftweave/internal/auth/adapter/redis"
	authhttp "github.com/raftweave/raftweave/internal/auth/handler/http"
	"github.com/raftweave/raftweave/internal/auth/handler/rpc"
	"github.com/raftweave/raftweave/internal/gen/auth/v1/authv1connect"
)

type Config struct {
	DBPool     *pgxpool.Pool
	Redis      *redis.Client
	Logger     *zap.Logger
	
	// JWT settings
	JWTPrivateKeyPEM []byte
	EncryptionKey    []byte
	
	// OAuth settings
	GitHubClientID     string
	GitHubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string
	
	// URLs
	CookieDomain string
	DashboardURL string
}

type Module struct {
	RPCHandlerPath string
	RPCHandler     http.Handler
	OAuthHandler   http.Handler
}

func Bootstrap(ctx context.Context, cfg Config) (*Module, error) {
	// 1. Adapters
	issuer, err := jwtadapter.New(cfg.JWTPrivateKeyPEM, "raftweave-auth", []string{"raftweave-dashboard"})
	if err != nil {
		return nil, fmt.Errorf("auth.di: %w", err)
	}

	tokenStore := redisadapter.NewRefreshTokenStore(cfg.Redis)
	userRepo := postgresadapter.NewUserRepository(cfg.DBPool)
	memberRepo := postgresadapter.NewMembershipRepository(cfg.DBPool)
	sessionRepo := postgresadapter.NewSessionRepository(cfg.DBPool)

	encryptor, err := crypto.NewEncryptor(cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("auth.di: %w", err)
	}

	stateStore := oauth.NewStateStore(cfg.Redis)
	ghProv := githubprovider.New(cfg.GitHubClientID, cfg.GitHubClientSecret, encryptor, stateStore, userRepo)
	ggProv := googleprovider.New(cfg.GoogleClientID, cfg.GoogleClientSecret, encryptor, stateStore, userRepo)

	// 2. RPC Handler
	// Note: We're using a simplified version of the constructor for now.
	// In a real app, you'd inject an actual OTP generator and Mailer.
	rpcHandler := rpc.NewAuthHandler(
		nil, // otpGen (placeholder)
		issuer,
		tokenStore,
		userRepo,
		memberRepo,
		sessionRepo,
		ghProv,
		nil, // mailer (placeholder)
		cfg.Logger,
	)

	// 3. OAuth Handler
	oauthHandler := authhttp.NewOAuthHandler(
		ghProv,
		ggProv,
		issuer,
		tokenStore,
		userRepo,
		memberRepo,
		cfg.CookieDomain,
		cfg.DashboardURL,
		cfg.Logger,
	)

	mux := http.NewServeMux()
	oauthHandler.RegisterRoutes(mux)

	interceptors := connect.WithInterceptors(rpc.NewAuthInterceptor(issuer))
	rpcPath, rpcSvc := authv1connect.NewAuthServiceHandler(rpcHandler, interceptors)

	return &Module{
		RPCHandlerPath: rpcPath,
		RPCHandler:     rpcSvc,
		OAuthHandler:   mux,
	}, nil
}

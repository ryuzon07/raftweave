package jwt

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/raftweave/raftweave/internal/auth/domain"
)

// Claims is the RaftWeave JWT payload.
// Follows RFC 7519 registered claims + RaftWeave-specific private claims.
type Claims struct {
	jwtlib.RegisteredClaims
	UserID    string            `json:"uid"`
	Email     string            `json:"email"`
	Name      string            `json:"name"`
	Provider  domain.Provider   `json:"provider"`
	SessionID string            `json:"sid"`
	Roles     map[string]string `json:"roles"`
	TokenType string            `json:"ttype"` // "access" | "refresh_meta"
}

// Issuer creates and validates JWTs.
type Issuer interface {
	IssueAccessToken(ctx context.Context, user *domain.User, sessionID string, roles map[string]string) (string, error)
	Validate(ctx context.Context, tokenStr string) (*Claims, error)
	PublicKey() *rsa.PublicKey
}

type rsaIssuer struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	audience   []string
}

// New creates a production Issuer using an RSA private key.
func New(privateKeyPEM []byte, issuer string, audience []string) (Issuer, error) {
	key, err := jwtlib.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("jwt.New: invalid private key: %w", err)
	}

	return &rsaIssuer{
		privateKey: key,
		publicKey:  &key.PublicKey,
		issuer:     issuer,
		audience:   audience,
	}, nil
}

func (i *rsaIssuer) IssueAccessToken(ctx context.Context, user *domain.User, sessionID string, roles map[string]string) (string, error) {
	now := time.Now().UTC()
	jti := uuid.New().String()

	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    i.issuer,
			Audience:  i.audience,
			ExpiresAt: jwtlib.NewNumericDate(now.Add(15 * time.Minute)),
			NotBefore: jwtlib.NewNumericDate(now),
			IssuedAt:  jwtlib.NewNumericDate(now),
			ID:        jti,
			Subject:   user.ID,
		},
		UserID:    user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Provider:  user.Provider,
		SessionID: sessionID,
		Roles:     roles,
		TokenType: "access",
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
	signed, err := token.SignedString(i.privateKey)
	if err != nil {
		return "", fmt.Errorf("jwt.IssueAccessToken: failed to sign token: %w", err)
	}

	return signed, nil
}

func (i *rsaIssuer) Validate(ctx context.Context, tokenStr string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenStr, &Claims{}, func(t *jwtlib.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return i.publicKey, nil
	}, jwtlib.WithIssuer(i.issuer), jwtlib.WithAudience(i.audience[0]), jwtlib.WithValidMethods([]string{"RS256"}))

	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.ErrTokenInvalid
	}

	if !token.Valid {
		return nil, domain.ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, domain.ErrTokenInvalid
	}

	if claims.TokenType != "access" {
		return nil, domain.ErrTokenInvalid
	}

	return claims, nil
}

func (i *rsaIssuer) PublicKey() *rsa.PublicKey {
	return i.publicKey
}

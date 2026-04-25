package domain

import "context"

// UserRepository is the persistence contract for User aggregates.
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByProviderID(ctx context.Context, provider Provider, providerID string) (*User, error)
	UpdateGitHubToken(ctx context.Context, userID string, encToken []byte) error
	UpdateLastLogin(ctx context.Context, userID string) error
	UpdateEmailVerified(ctx context.Context, userID string) error
	SoftDelete(ctx context.Context, userID string) error
}

// SessionRepository persists refresh token sessions.
type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	GetByID(ctx context.Context, id string) (*Session, error)
	GetByUserID(ctx context.Context, userID string) ([]*Session, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) (int64, error)
}

// OTPRepository manages short-lived OTP challenges (backed by Redis).
type OTPRepository interface {
	Create(ctx context.Context, challenge *OTPChallenge) error
	GetByID(ctx context.Context, id string) (*OTPChallenge, error)
	IncrementAttempts(ctx context.Context, id string) (int, error)
	MarkUsed(ctx context.Context, id string) error
	DeleteByEmail(ctx context.Context, email string) error // invalidates prior OTPs
}

// MembershipRepository manages workspace role assignments.
type MembershipRepository interface {
	Upsert(ctx context.Context, m *WorkspaceMembership) error
	GetByUserAndWorkspace(ctx context.Context, userID, workspaceID string) (*WorkspaceMembership, error)
	GetUserWorkspaces(ctx context.Context, userID string) ([]*WorkspaceMembership, error)
	RemoveMember(ctx context.Context, userID, workspaceID string) error
}

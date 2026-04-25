package domain

import (
	"errors"
	"time"
)

// Provider identifies the authentication method used.
type Provider string

const (
	ProviderGitHub Provider = "github"
	ProviderGoogle Provider = "google"
	ProviderEmail  Provider = "email"
)

// Role defines what a user is permitted to do within a workspace.
type Role string

const (
	RoleOwner  Role = "OWNER"  // full control including billing and deletion
	RoleAdmin  Role = "ADMIN"  // can manage members, provision resources
	RoleMember Role = "MEMBER" // can deploy workloads, view logs
	RoleViewer Role = "VIEWER" // read-only access to dashboard and logs
)

// User is the aggregate root for an authenticated identity.
type User struct {
	ID              string
	Email           string
	Name            string
	AvatarURL       string
	Provider        Provider
	ProviderID      string // GitHub user ID, Google sub, etc.
	GitHubLogin     string // GitHub username — used for repo operations
	GitHubTokenEnc  []byte // AES-256-GCM encrypted GitHub OAuth token
	IsEmailVerified bool
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastLoginAt     *time.Time
}

// Session represents an active authenticated session.
type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string // bcrypt hash of the opaque refresh token
	Fingerprint      string // SHA-256(User-Agent + IP) — detects token theft
	ExpiresAt        time.Time
	CreatedAt        time.Time
	RevokedAt        *time.Time
	IsRevoked        bool
}

// OTPChallenge is a pending email OTP verification.
type OTPChallenge struct {
	ID        string
	Email     string
	CodeHash  string // bcrypt hash of the 6-digit OTP
	Purpose   OTPPurpose
	Attempts  int
	ExpiresAt time.Time
	CreatedAt time.Time
	Used      bool
}

// OTPPurpose indicates the purpose of an OTP challenge.
type OTPPurpose string

const (
	OTPPurposeLogin        OTPPurpose = "LOGIN"
	OTPPurposeEmailVerify  OTPPurpose = "EMAIL_VERIFY"
	OTPPurposePasswordless OTPPurpose = "PASSWORDLESS"
)

// WorkspaceMembership links a user to a workspace with a role.
type WorkspaceMembership struct {
	UserID      string
	WorkspaceID string
	Role        Role
	InvitedBy   string
	JoinedAt    time.Time
}

// Domain errors — wrap these; never expose raw infrastructure errors.
var (
	ErrUserNotFound               = errors.New("user not found")
	ErrUserAlreadyExists          = errors.New("user already exists with this email")
	ErrSessionNotFound            = errors.New("session not found")
	ErrSessionExpired             = errors.New("session expired")
	ErrSessionRevoked             = errors.New("session revoked")
	ErrSessionFingerprintMismatch = errors.New("session fingerprint mismatch — possible token theft")
	ErrOTPNotFound                = errors.New("otp challenge not found")
	ErrOTPExpired                 = errors.New("otp has expired")
	ErrOTPInvalid                 = errors.New("otp code is incorrect")
	ErrOTPMaxAttemptsReached      = errors.New("max otp verification attempts reached")
	ErrOTPAlreadyUsed             = errors.New("otp has already been used")
	ErrProviderMismatch           = errors.New("email already registered with a different provider")
	ErrInsufficientRole           = errors.New("user does not have required role")
	ErrTokenInvalid               = errors.New("token is invalid or malformed")
	ErrTokenExpired               = errors.New("token has expired")
)

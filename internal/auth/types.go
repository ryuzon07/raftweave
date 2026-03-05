package auth

import "time"

// User represents an authenticated user.
type User struct {
	ID        string
	Email     string
	Name      string
	AvatarURL string
	Provider  string // github | google
	CreatedAt time.Time
}

// Session represents an active user session.
type Session struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Role represents a user's authorization role.
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleViewer Role = "viewer"
)

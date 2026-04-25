// Package postgres implements domain repository interfaces using PostgreSQL via pgx.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/raftweave/raftweave/internal/auth/domain"
)

// UserRepository implements domain.UserRepository with PostgreSQL.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new PostgreSQL-backed user repository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, name, avatar_url, provider, provider_id, github_login, github_token_enc,
		                     is_email_verified, is_active, created_at, updated_at, last_login_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		u.ID, u.Email, u.Name, u.AvatarURL, string(u.Provider), u.ProviderID,
		u.GitHubLogin, u.GitHubTokenEnc, u.IsEmailVerified, u.IsActive,
		u.CreatedAt, u.UpdatedAt, u.LastLoginAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("userRepo.Create: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return r.scanUser(ctx, `SELECT * FROM users WHERE id = $1 AND is_active = TRUE`, id)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.scanUser(ctx, `SELECT * FROM users WHERE LOWER(email) = LOWER($1) AND is_active = TRUE`, email)
}

func (r *UserRepository) GetByProviderID(ctx context.Context, provider domain.Provider, providerID string) (*domain.User, error) {
	return r.scanUser(ctx,
		`SELECT * FROM users WHERE provider = $1 AND provider_id = $2 AND is_active = TRUE`,
		string(provider), providerID,
	)
}

func (r *UserRepository) UpdateGitHubToken(ctx context.Context, userID string, encToken []byte) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET github_token_enc = $1, updated_at = $2 WHERE id = $3`,
		encToken, time.Now().UTC(), userID,
	)
	if err != nil {
		return fmt.Errorf("userRepo.UpdateGitHubToken: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3`,
		now, now, userID,
	)
	if err != nil {
		return fmt.Errorf("userRepo.UpdateLastLogin: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateEmailVerified(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET is_email_verified = TRUE, updated_at = $1 WHERE id = $2`,
		time.Now().UTC(), userID,
	)
	if err != nil {
		return fmt.Errorf("userRepo.UpdateEmailVerified: %w", err)
	}
	return nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET is_active = FALSE, updated_at = $1 WHERE id = $2`,
		time.Now().UTC(), userID,
	)
	if err != nil {
		return fmt.Errorf("userRepo.SoftDelete: %w", err)
	}
	return nil
}

func (r *UserRepository) scanUser(ctx context.Context, query string, args ...interface{}) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	u := &domain.User{}
	var providerStr string
	err := row.Scan(
		&u.ID, &u.Email, &u.Name, &u.AvatarURL,
		&providerStr, &u.ProviderID, &u.GitHubLogin, &u.GitHubTokenEnc,
		&u.IsEmailVerified, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("userRepo.scanUser: %w", err)
	}
	u.Provider = domain.Provider(strings.ToLower(providerStr))
	return u, nil
}

// SessionRepository implements domain.SessionRepository with PostgreSQL.
type SessionRepository struct {
	pool *pgxpool.Pool
}

// NewSessionRepository creates a new PostgreSQL-backed session repository.
func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sessions (id, user_id, refresh_token_hash, fingerprint, expires_at, created_at, is_revoked)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		s.ID, s.UserID, s.RefreshTokenHash, s.Fingerprint, s.ExpiresAt, s.CreatedAt, s.IsRevoked,
	)
	if err != nil {
		return fmt.Errorf("sessionRepo.Create: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, user_id, refresh_token_hash, fingerprint, expires_at, created_at, revoked_at, is_revoked FROM sessions WHERE id = $1`, id)
	s := &domain.Session{}
	err := row.Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.Fingerprint, &s.ExpiresAt, &s.CreatedAt, &s.RevokedAt, &s.IsRevoked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) { return nil, domain.ErrSessionNotFound }
		return nil, fmt.Errorf("sessionRepo.GetByID: %w", err)
	}
	return s, nil
}

func (r *SessionRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, refresh_token_hash, fingerprint, expires_at, created_at, revoked_at, is_revoked
		 FROM sessions WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("sessionRepo.GetByUserID: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		s := &domain.Session{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.Fingerprint, &s.ExpiresAt, &s.CreatedAt, &s.RevokedAt, &s.IsRevoked); err != nil {
			return nil, fmt.Errorf("sessionRepo.GetByUserID scan: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, id string) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE sessions SET is_revoked = TRUE, revoked_at = $1 WHERE id = $2`, now, id)
	if err != nil {
		return fmt.Errorf("sessionRepo.Revoke: %w", err)
	}
	return nil
}

func (r *SessionRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE sessions SET is_revoked = TRUE, revoked_at = $1 WHERE user_id = $2 AND is_revoked = FALSE`, now, userID)
	if err != nil {
		return fmt.Errorf("sessionRepo.RevokeAllForUser: %w", err)
	}
	return nil
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < NOW()`)
	if err != nil {
		return 0, fmt.Errorf("sessionRepo.DeleteExpired: %w", err)
	}
	return tag.RowsAffected(), nil
}

// MembershipRepository implements domain.MembershipRepository with PostgreSQL.
type MembershipRepository struct {
	pool *pgxpool.Pool
}

// NewMembershipRepository creates a new PostgreSQL-backed membership repository.
func NewMembershipRepository(pool *pgxpool.Pool) *MembershipRepository {
	return &MembershipRepository{pool: pool}
}

func (r *MembershipRepository) Upsert(ctx context.Context, m *domain.WorkspaceMembership) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO workspace_memberships (user_id, workspace_id, role, invited_by, joined_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id, workspace_id) DO UPDATE SET role = $3`,
		m.UserID, m.WorkspaceID, string(m.Role), m.InvitedBy, m.JoinedAt,
	)
	if err != nil {
		return fmt.Errorf("membershipRepo.Upsert: %w", err)
	}
	return nil
}

func (r *MembershipRepository) GetByUserAndWorkspace(ctx context.Context, userID, workspaceID string) (*domain.WorkspaceMembership, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT user_id, workspace_id, role, invited_by, joined_at FROM workspace_memberships WHERE user_id = $1 AND workspace_id = $2`,
		userID, workspaceID)
	m := &domain.WorkspaceMembership{}
	var roleStr string
	if err := row.Scan(&m.UserID, &m.WorkspaceID, &roleStr, &m.InvitedBy, &m.JoinedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) { return nil, domain.ErrUserNotFound }
		return nil, fmt.Errorf("membershipRepo.GetByUserAndWorkspace: %w", err)
	}
	m.Role = domain.Role(roleStr)
	return m, nil
}

func (r *MembershipRepository) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.WorkspaceMembership, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT user_id, workspace_id, role, invited_by, joined_at FROM workspace_memberships WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("membershipRepo.GetUserWorkspaces: %w", err)
	}
	defer rows.Close()

	var memberships []*domain.WorkspaceMembership
	for rows.Next() {
		m := &domain.WorkspaceMembership{}
		var roleStr string
		if err := rows.Scan(&m.UserID, &m.WorkspaceID, &roleStr, &m.InvitedBy, &m.JoinedAt); err != nil {
			return nil, err
		}
		m.Role = domain.Role(roleStr)
		memberships = append(memberships, m)
	}
	return memberships, nil
}

func (r *MembershipRepository) RemoveMember(ctx context.Context, userID, workspaceID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM workspace_memberships WHERE user_id = $1 AND workspace_id = $2`, userID, workspaceID)
	if err != nil {
		return fmt.Errorf("membershipRepo.RemoveMember: %w", err)
	}
	return nil
}

// Package otp implements passwordless email OTP login with rate limiting.
package otp

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/raftweave/raftweave/internal/auth/domain"
)

const (
	// OTPLength is the number of digits in the OTP.
	OTPLength = 6
	// OTPMaxAttempts is the max verification attempts before challenge deletion.
	OTPMaxAttempts = 3
	// OTPTTL is the OTP validity duration in seconds.
	OTPTTL = 10 * 60
	// bcryptCost for OTP hashing.
	bcryptCost = 10
)

// Generator creates and verifies OTP challenges.
type Generator interface {
	// Issue creates a new OTP challenge, stores it, and sends the email.
	Issue(ctx context.Context, email string, purpose domain.OTPPurpose) (challengeID string, err error)
	// Verify checks the OTP code against the stored challenge.
	Verify(ctx context.Context, challengeID, code string) (email string, err error)
}

// RateLimiter checks whether an action is rate-limited.
type RateLimiter interface {
	// Allow returns true if the action is allowed; false if rate-limited.
	Allow(ctx context.Context, key string) (bool, error)
}

type generator struct {
	otpRepo     domain.OTPRepository
	mailer      Mailer
	rateLimiter RateLimiter
}

// NewGenerator creates a new OTP generator.
func NewGenerator(repo domain.OTPRepository, mailer Mailer, rl RateLimiter) Generator {
	return &generator{otpRepo: repo, mailer: mailer, rateLimiter: rl}
}

func (g *generator) Issue(ctx context.Context, email string, purpose domain.OTPPurpose) (string, error) {
	// Rate limit check.
	if g.rateLimiter != nil {
		allowed, err := g.rateLimiter.Allow(ctx, "otp:"+email)
		if err != nil {
			return "", fmt.Errorf("otp.Issue: rate limiter: %w", err)
		}
		if !allowed {
			// SECURITY: return same response shape to prevent user enumeration.
			return uuid.New().String(), nil
		}
	}

	// Invalidate prior OTPs for this email.
	_ = g.otpRepo.DeleteByEmail(ctx, email)

	// Generate cryptographically random 6-digit OTP.
	code, err := generateNumericOTP(OTPLength)
	if err != nil {
		return "", fmt.Errorf("otp.Issue: generate: %w", err)
	}

	// Hash with bcrypt.
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("otp.Issue: hash: %w", err)
	}

	challengeID := uuid.New().String()
	now := time.Now().UTC()
	challenge := &domain.OTPChallenge{
		ID:        challengeID,
		Email:     email,
		CodeHash:  string(hash),
		Purpose:   purpose,
		Attempts:  0,
		ExpiresAt: now.Add(time.Duration(OTPTTL) * time.Second),
		CreatedAt: now,
		Used:      false,
	}

	if err := g.otpRepo.Create(ctx, challenge); err != nil {
		return "", fmt.Errorf("otp.Issue: store: %w", err)
	}

	// Send email (non-blocking error — still return challenge ID).
	if g.mailer != nil {
		if err := g.mailer.SendOTP(ctx, email, code, challenge.ExpiresAt); err != nil {
			// Log but don't fail — the challenge is already stored.
			_ = err
		}
	}

	return challengeID, nil
}

func (g *generator) Verify(ctx context.Context, challengeID, code string) (string, error) {
	challenge, err := g.otpRepo.GetByID(ctx, challengeID)
	if err != nil {
		return "", domain.ErrOTPNotFound
	}

	if challenge.Used {
		return "", domain.ErrOTPAlreadyUsed
	}

	if time.Now().UTC().After(challenge.ExpiresAt) {
		return "", domain.ErrOTPExpired
	}

	if challenge.Attempts >= OTPMaxAttempts {
		_ = g.otpRepo.DeleteByEmail(ctx, challenge.Email)
		return "", domain.ErrOTPMaxAttemptsReached
	}

	// Compare code using bcrypt (constant-time).
	if err := bcrypt.CompareHashAndPassword([]byte(challenge.CodeHash), []byte(code)); err != nil {
		attempts, _ := g.otpRepo.IncrementAttempts(ctx, challengeID)
		if attempts >= OTPMaxAttempts {
			_ = g.otpRepo.DeleteByEmail(ctx, challenge.Email)
			return "", domain.ErrOTPMaxAttemptsReached
		}
		return "", domain.ErrOTPInvalid
	}

	// Mark as used.
	if err := g.otpRepo.MarkUsed(ctx, challengeID); err != nil {
		return "", fmt.Errorf("otp.Verify: mark used: %w", err)
	}

	return challenge.Email, nil
}

// generateNumericOTP generates a cryptographically random N-digit numeric string.
func generateNumericOTP(digits int) (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("generateNumericOTP: %w", err)
	}
	return fmt.Sprintf("%0*d", digits, n), nil
}

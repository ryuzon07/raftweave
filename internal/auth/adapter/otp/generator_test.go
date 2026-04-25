package otp

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/raftweave/raftweave/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// In-memory OTP repository for testing.
type memOTPRepo struct {
	mu         sync.Mutex
	challenges map[string]*domain.OTPChallenge
}

func newMemOTPRepo() *memOTPRepo {
	return &memOTPRepo{challenges: make(map[string]*domain.OTPChallenge)}
}

func (m *memOTPRepo) Create(_ context.Context, c *domain.OTPChallenge) error {
	m.mu.Lock(); defer m.mu.Unlock()
	m.challenges[c.ID] = c; return nil
}

func (m *memOTPRepo) GetByID(_ context.Context, id string) (*domain.OTPChallenge, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	c, ok := m.challenges[id]
	if !ok { return nil, domain.ErrOTPNotFound }
	return c, nil
}

func (m *memOTPRepo) IncrementAttempts(_ context.Context, id string) (int, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	c, ok := m.challenges[id]
	if !ok { return 0, domain.ErrOTPNotFound }
	c.Attempts++; return c.Attempts, nil
}

func (m *memOTPRepo) MarkUsed(_ context.Context, id string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	c, ok := m.challenges[id]
	if !ok { return domain.ErrOTPNotFound }
	c.Used = true; return nil
}

func (m *memOTPRepo) DeleteByEmail(_ context.Context, email string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	for id, c := range m.challenges {
		if c.Email == email { delete(m.challenges, id) }
	}
	return nil
}

// allowAllLimiter always allows.
type allowAllLimiter struct{}
func (allowAllLimiter) Allow(_ context.Context, _ string) (bool, error) { return true, nil }

// countingLimiter tracks calls and blocks after a threshold.
type countingLimiter struct {
	mu    sync.Mutex
	calls map[string]int
	max   int
}

func newCountingLimiter(max int) *countingLimiter {
	return &countingLimiter{calls: make(map[string]int), max: max}
}

func (l *countingLimiter) Allow(_ context.Context, key string) (bool, error) {
	l.mu.Lock(); defer l.mu.Unlock()
	l.calls[key]++
	return l.calls[key] <= l.max, nil
}

// capturingMailer records sent OTPs for testing.
type capturingMailer struct {
	mu    sync.Mutex
	codes []string
}

func (m *capturingMailer) SendOTP(_ context.Context, _, code string, _ time.Time) error {
	m.mu.Lock(); defer m.mu.Unlock()
	m.codes = append(m.codes, code); return nil
}
func (m *capturingMailer) SendSecurityAlert(_ context.Context, _, _ string) error { return nil }
func (m *capturingMailer) SendAccountLinked(_ context.Context, _, _ string) error { return nil }

func TestIssue_GeneratesUniqueChallengeIDs(t *testing.T) {
	repo := newMemOTPRepo()
	gen := NewGenerator(repo, NoopMailer{}, allowAllLimiter{})
	ids := make(map[string]struct{})
	for i := 0; i < 50; i++ {
		id, err := gen.Issue(context.Background(), "test@example.com", domain.OTPPurposeLogin)
		require.NoError(t, err)
		_, exists := ids[id]; assert.False(t, exists)
		ids[id] = struct{}{}
	}
}

func TestIssue_OTPIsNumericAndCorrectLength(t *testing.T) {
	repo := newMemOTPRepo()
	mailer := &capturingMailer{}
	gen := NewGenerator(repo, mailer, allowAllLimiter{})
	_, err := gen.Issue(context.Background(), "test@example.com", domain.OTPPurposeLogin)
	require.NoError(t, err)
	require.Len(t, mailer.codes, 1)
	code := mailer.codes[0]
	assert.Len(t, code, OTPLength)
	_, err = strconv.Atoi(code)
	assert.NoError(t, err, "OTP should be numeric")
}

func TestVerify_CorrectCode_ReturnsEmail(t *testing.T) {
	repo := newMemOTPRepo()
	mailer := &capturingMailer{}
	gen := NewGenerator(repo, mailer, allowAllLimiter{})
	cid, err := gen.Issue(context.Background(), "user@test.com", domain.OTPPurposeLogin)
	require.NoError(t, err)
	email, err := gen.Verify(context.Background(), cid, mailer.codes[0])
	require.NoError(t, err)
	assert.Equal(t, "user@test.com", email)
}

func TestVerify_IncorrectCode_IncrementsAttempts(t *testing.T) {
	repo := newMemOTPRepo()
	gen := NewGenerator(repo, &capturingMailer{}, allowAllLimiter{})
	cid, _ := gen.Issue(context.Background(), "user@test.com", domain.OTPPurposeLogin)
	_, err := gen.Verify(context.Background(), cid, "000000")
	assert.ErrorIs(t, err, domain.ErrOTPInvalid)
	ch, _ := repo.GetByID(context.Background(), cid)
	assert.Equal(t, 1, ch.Attempts)
}

func TestVerify_MaxAttempts_DeletesChallenge(t *testing.T) {
	repo := newMemOTPRepo()
	gen := NewGenerator(repo, &capturingMailer{}, allowAllLimiter{})
	cid, _ := gen.Issue(context.Background(), "user@test.com", domain.OTPPurposeLogin)
	for i := 0; i < OTPMaxAttempts; i++ {
		_, _ = gen.Verify(context.Background(), cid, "000000")
	}
	_, err := gen.Verify(context.Background(), cid, "000000")
	assert.ErrorIs(t, err, domain.ErrOTPNotFound)
}

func TestVerify_ExpiredChallenge_ReturnsErrOTPExpired(t *testing.T) {
	repo := newMemOTPRepo()
	mailer := &capturingMailer{}
	gen := NewGenerator(repo, mailer, allowAllLimiter{})
	cid, _ := gen.Issue(context.Background(), "user@test.com", domain.OTPPurposeLogin)
	// Manually expire the challenge.
	ch, _ := repo.GetByID(context.Background(), cid)
	ch.ExpiresAt = time.Now().UTC().Add(-1 * time.Minute)
	_, err := gen.Verify(context.Background(), cid, mailer.codes[0])
	assert.ErrorIs(t, err, domain.ErrOTPExpired)
}

func TestVerify_AlreadyUsedChallenge_ReturnsErrOTPAlreadyUsed(t *testing.T) {
	repo := newMemOTPRepo()
	mailer := &capturingMailer{}
	gen := NewGenerator(repo, mailer, allowAllLimiter{})
	cid, _ := gen.Issue(context.Background(), "user@test.com", domain.OTPPurposeLogin)
	_, err := gen.Verify(context.Background(), cid, mailer.codes[0])
	require.NoError(t, err)
	_, err = gen.Verify(context.Background(), cid, mailer.codes[0])
	assert.ErrorIs(t, err, domain.ErrOTPAlreadyUsed)
}

func TestIssue_PriorOTPInvalidated_OnNewRequest(t *testing.T) {
	repo := newMemOTPRepo()
	mailer := &capturingMailer{}
	gen := NewGenerator(repo, mailer, allowAllLimiter{})
	cid1, _ := gen.Issue(context.Background(), "user@test.com", domain.OTPPurposeLogin)
	_, _ = gen.Issue(context.Background(), "user@test.com", domain.OTPPurposeLogin)
	// First challenge should be deleted.
	_, err := gen.Verify(context.Background(), cid1, mailer.codes[0])
	assert.ErrorIs(t, err, domain.ErrOTPNotFound)
}

func TestIssue_RateLimit_BlocksAfterThreeRequests(t *testing.T) {
	repo := newMemOTPRepo()
	rl := newCountingLimiter(3)
	gen := NewGenerator(repo, NoopMailer{}, rl)
	for i := 0; i < 3; i++ {
		_, err := gen.Issue(context.Background(), "user@test.com", domain.OTPPurposeLogin)
		require.NoError(t, err)
	}
	// 4th request should be rate-limited (returns a dummy ID, no challenge stored).
	cid, err := gen.Issue(context.Background(), "user@test.com", domain.OTPPurposeLogin)
	require.NoError(t, err)
	assert.NotEmpty(t, cid, "should still return an ID to prevent user enumeration")
	// But verifying it should fail.
	_, err = gen.Verify(context.Background(), cid, "123456")
	assert.Error(t, err)
}

func TestGenerateNumericOTP_CorrectLength(t *testing.T) {
	for i := 0; i < 100; i++ {
		code, err := generateNumericOTP(6)
		require.NoError(t, err)
		assert.Len(t, code, 6)
		_, err = strconv.Atoi(code)
		assert.NoError(t, err)
	}
}

func TestOTPHashUsesBcrypt(t *testing.T) {
	code := "123456"
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcryptCost)
	require.NoError(t, err)
	// Verify cost >= 10
	cost, err := bcrypt.Cost(hash)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, cost, 10)
}

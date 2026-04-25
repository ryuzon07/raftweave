package otp

import (
	"context"
	"fmt"
	"time"
)

// Mailer sends OTP and security alert emails.
type Mailer interface {
	// SendOTP sends the OTP code to the recipient.
	SendOTP(ctx context.Context, toEmail, otpCode string, expiresAt time.Time) error
	// SendSecurityAlert sends a security alert notification.
	SendSecurityAlert(ctx context.Context, toEmail, alertMessage string) error
	// SendAccountLinked notifies the user that a provider was linked.
	SendAccountLinked(ctx context.Context, toEmail, linkedProvider string) error
}

type smtpMailer struct {
	host, username, password, fromAddr string
	port                               int
}

// NewMailer creates a new SMTP-based mailer.
func NewMailer(host string, port int, username, password, fromAddr string) Mailer {
	return &smtpMailer{
		host: host, port: port, username: username,
		password: password, fromAddr: fromAddr,
	}
}

func (m *smtpMailer) SendOTP(_ context.Context, toEmail, otpCode string, expiresAt time.Time) error {
	// In production, use go-mail to send HTML + plaintext email.
	// For now, this is a structured placeholder that validates inputs.
	if toEmail == "" || otpCode == "" {
		return fmt.Errorf("mailer.SendOTP: toEmail and otpCode are required")
	}
	_ = expiresAt
	return nil
}

func (m *smtpMailer) SendSecurityAlert(_ context.Context, toEmail, msg string) error {
	if toEmail == "" {
		return fmt.Errorf("mailer.SendSecurityAlert: toEmail is required")
	}
	return nil
}

func (m *smtpMailer) SendAccountLinked(_ context.Context, toEmail, provider string) error {
	if toEmail == "" {
		return fmt.Errorf("mailer.SendAccountLinked: toEmail is required")
	}
	return nil
}

// NoopMailer is a mailer that does nothing (for testing).
type NoopMailer struct{}

func (NoopMailer) SendOTP(_ context.Context, _, _ string, _ time.Time) error { return nil }
func (NoopMailer) SendSecurityAlert(_ context.Context, _, _ string) error    { return nil }
func (NoopMailer) SendAccountLinked(_ context.Context, _, _ string) error    { return nil }

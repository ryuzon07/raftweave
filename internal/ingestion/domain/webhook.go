package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type WebhookProvider string

const (
	WebhookProviderGitHub WebhookProvider = "github"
	WebhookProviderGitLab WebhookProvider = "gitlab"
)

type WebhookEvent struct {
	ID           string
	Provider     WebhookProvider
	WorkloadName string
	RepoURL      string
	Branch       string
	CommitSHA    string
	CommitMsg    string
	PushedBy     string
	TriggeredAt  time.Time
	Signature    string
	RawPayload   []byte
}

func (e *WebhookEvent) VerifySignature(secret string) error {
	if e.Signature == "" {
		return ErrInvalidSignature
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(e.RawPayload)
	expectedMAC := mac.Sum(nil)

	var providedMAC []byte
	var err error

	if e.Provider == WebhookProviderGitHub {
		// GitHub signature format is "sha256=..."
		if len(e.Signature) > 7 && e.Signature[:7] == "sha256=" {
			providedMAC, err = hex.DecodeString(e.Signature[7:])
		} else {
			return ErrInvalidSignature
		}
	} else if e.Provider == WebhookProviderGitLab {
		// GitLab typically sends plain text token or HMAC, adjust accordingly if needed
		// For consistency in this domain, assuming hex encoded HMAC
		providedMAC, err = hex.DecodeString(e.Signature)
	} else {
		return ErrInvalidSignature
	}

	if err != nil {
		return ErrInvalidSignature
	}

	if !hmac.Equal(providedMAC, expectedMAC) {
		return ErrInvalidSignature
	}

	return nil
}

func (e *WebhookEvent) IsDefaultBranch(configuredBranch string) bool {
	return e.Branch == configuredBranch
}

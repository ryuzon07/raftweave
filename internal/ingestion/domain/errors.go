package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrWorkloadNotFound      = errors.New("workload not found")
	ErrWorkloadAlreadyExists = errors.New("workload with this name already exists")
	ErrInvalidWorkloadName   = errors.New("workload name must match ^[a-z][a-z0-9-]{2,62}$")
	ErrInvalidRegion         = errors.New("invalid cloud region")
	ErrInvalidSource         = errors.New("source must specify either git repo or image ref")
	ErrInvalidRTO            = errors.New("rto_seconds must be between 10 and 300")
	ErrInvalidRPO            = errors.New("rpo_seconds must be between 1 and 60")
	ErrCredentialNotFound    = errors.New("credential not found")
	ErrInvalidSignature      = errors.New("webhook signature validation failed")
	ErrMissingCredential     = errors.New("no credential found for provider")
	ErrEncryptionFailed      = errors.New("credential encryption failed")
	ErrDecryptionFailed      = errors.New("credential decryption failed")
)

type ValidationError struct {
	Field   string
	Message string
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var msgs []string
	for _, err := range v {
		msgs = append(msgs, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(msgs, "; ")
}

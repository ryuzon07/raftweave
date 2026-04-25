package domain

import (
	"time"

	"github.com/google/uuid"
)

type CredentialID string

type CredentialType string

const (
	CredentialTypeAWSIAM            CredentialType = "aws_iam"
	CredentialTypeAzureServicePrincipal CredentialType = "azure_sp"
	CredentialTypeGCPServiceAccount CredentialType = "gcp_sa"
)

type CloudCredential struct {
	ID               CredentialID
	WorkloadID       WorkloadID
	Provider         CloudProvider
	Type             CredentialType
	EncryptedPayload []byte
	KeyVersion       string
	CreatedAt        time.Time
	RotatedAt        *time.Time
}

func NewCloudCredential(workloadID WorkloadID, provider CloudProvider, credType CredentialType, rawPayload []byte, encryptor Encryptor) (*CloudCredential, error) {
	if len(rawPayload) == 0 {
		return nil, ErrMissingCredential
	}

	ciphertext, keyVersion, err := encryptor.Encrypt(rawPayload)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	return &CloudCredential{
		ID:               CredentialID(uuid.New().String()),
		WorkloadID:       workloadID,
		Provider:         provider,
		Type:             credType,
		EncryptedPayload: ciphertext,
		KeyVersion:       keyVersion,
		CreatedAt:        time.Now().UTC(),
	}, nil
}

func (c *CloudCredential) Decrypt(encryptor Encryptor) ([]byte, error) {
	plaintext, err := encryptor.Decrypt(c.EncryptedPayload, c.KeyVersion)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	return plaintext, nil
}

func (c *CloudCredential) Rotate(newPayload []byte, encryptor Encryptor) error {
	if len(newPayload) == 0 {
		return ErrMissingCredential
	}

	ciphertext, keyVersion, err := encryptor.Encrypt(newPayload)
	if err != nil {
		return ErrEncryptionFailed
	}

	c.EncryptedPayload = ciphertext
	c.KeyVersion = keyVersion
	now := time.Now().UTC()
	c.RotatedAt = &now
	return nil
}

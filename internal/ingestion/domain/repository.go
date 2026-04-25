package domain

import (
	"context"
)

type WorkloadRepository interface {
	Save(ctx context.Context, w *Workload) error
	FindByID(ctx context.Context, id WorkloadID) (*Workload, error)
	FindByName(ctx context.Context, name string) (*Workload, error)
	FindAll(ctx context.Context, limit, offset int) ([]*Workload, error)
	UpdateStatus(ctx context.Context, id WorkloadID, status WorkloadStatus) error
	Delete(ctx context.Context, id WorkloadID) error
}

type CredentialRepository interface {
	Save(ctx context.Context, c *CloudCredential) error
	FindByWorkloadAndProvider(ctx context.Context, workloadID WorkloadID, provider CloudProvider) (*CloudCredential, error)
	FindAllByWorkload(ctx context.Context, workloadID WorkloadID) ([]*CloudCredential, error)
	Delete(ctx context.Context, id CredentialID) error
}

type Encryptor interface {
	Encrypt(plaintext []byte) (ciphertext []byte, keyVersion string, err error)
	Decrypt(ciphertext []byte, keyVersion string) (plaintext []byte, err error)
}

type JobEnqueuer interface {
	EnqueueBuildJob(ctx context.Context, workloadID WorkloadID, event *WebhookEvent) error
	EnqueueProvisionJob(ctx context.Context, workloadID WorkloadID) error
}

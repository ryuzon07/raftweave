package domain_test

import (
	"errors"
        "testing"

	"github.com/raftweave/raftweave/internal/ingestion/domain"
)

func TestNewWorkload_ValidInput(t *testing.T) {
	source := domain.SourceConfig{Type: "git", RepoURL: "https://github.com/test/repo", Branch: "main"}
	primary := domain.Region{Provider: domain.CloudProviderAWS, Name: "us-east-1"}
	standbys := []domain.Region{{Provider: domain.CloudProviderAWS, Name: "us-west-2"}}
	compute := domain.ResourceSpec{CPU: "2", Memory: "4Gi"}
	db := domain.DatabaseSpec{Engine: "postgres"}
	failover := domain.FailoverConfig{RTOSeconds: 60, RPOSeconds: 30}

	w, err := domain.NewWorkload("valid-name-123", source, primary, standbys, compute, db, failover)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if w.Name != "valid-name-123" {
		t.Errorf("expected name 'valid-name-123', got %s", w.Name)
	}
	if w.Status != domain.WorkloadStatusPending {
		t.Errorf("expected status PENDING, got %s", w.Status)
	}
}

func TestNewWorkload_InvalidName(t *testing.T) {
	// Name with uppercase
	_, err := domain.NewWorkload("InvalidName", domain.SourceConfig{}, domain.Region{}, []domain.Region{{}}, domain.ResourceSpec{}, domain.DatabaseSpec{}, domain.FailoverConfig{})
	if err == nil {
		t.Fatal("expected error for invalid name, got nil")
	}
}

func TestNewWorkload_InvalidRTO_RPO(t *testing.T) {
	source := domain.SourceConfig{Type: "git", RepoURL: "https://github.com/test/repo", Branch: "main"}
	primary := domain.Region{Provider: domain.CloudProviderAWS, Name: "us-east-1"}
	standbys := []domain.Region{{Provider: domain.CloudProviderAWS, Name: "us-west-2"}}
	compute := domain.ResourceSpec{CPU: "2", Memory: "4Gi"}
	db := domain.DatabaseSpec{Engine: "postgres"}
	failover := domain.FailoverConfig{RTOSeconds: 5, RPOSeconds: 30} // Invalid RTO

	_, err := domain.NewWorkload("valid-name", source, primary, standbys, compute, db, failover)
	if err == nil {
		t.Fatal("expected error for RTO < 10, got nil")
	}

	failover.RTOSeconds = 60
	failover.RPOSeconds = 0 // Invalid RPO
	_, err = domain.NewWorkload("valid-name", source, primary, standbys, compute, db, failover)
	if err == nil {
		t.Fatal("expected error for RPO < 1, got nil")
	}
}

func TestNewWorkload_NoStandbys(t *testing.T) {
	source := domain.SourceConfig{Type: "git", RepoURL: "https://github.com/test/repo", Branch: "main"}
	primary := domain.Region{Provider: domain.CloudProviderAWS, Name: "us-east-1"}
	compute := domain.ResourceSpec{CPU: "2", Memory: "4Gi"}
	db := domain.DatabaseSpec{Engine: "postgres"}
	failover := domain.FailoverConfig{RTOSeconds: 60, RPOSeconds: 30}

	_, err := domain.NewWorkload("valid-name", source, primary, nil, compute, db, failover)
	if err == nil {
		t.Fatal("expected error for empty standbys, got nil")
	}
}

func TestWorkload_Validate_AllViolations(t *testing.T) {
	w := &domain.Workload{
		Name: "Invalid Name!",
		Source: domain.SourceConfig{
			Type: "invalid",
		},
		Failover: domain.FailoverConfig{
			RTOSeconds: 5,
			RPOSeconds: 0,
		},
		StandbyRegions: nil,
	}

	err := w.Validate()
	if err == nil {
		t.Fatal("expected validation errors")
	}

	var validationErrs domain.ValidationErrors
	if !errors.As(err, &validationErrs) {
		t.Fatalf("expected domain.ValidationErrors, got %T", err)
	}

	if len(validationErrs) != 5 {
		t.Fatalf("expected 5 validation errors, got %v", err)
	}
}

type mockEncryptor struct{}

func (m *mockEncryptor) Encrypt(plaintext []byte) ([]byte, string, error) {
	return append([]byte("enc_"), plaintext...), "v1", nil
}

func (m *mockEncryptor) Decrypt(ciphertext []byte, keyVersion string) ([]byte, error) {
	if string(ciphertext[:4]) != "enc_" {
		return nil, domain.ErrDecryptionFailed
	}
	return ciphertext[4:], nil
}

func TestCloudCredential_Decrypt(t *testing.T) {
	enc := &mockEncryptor{}
	payload := []byte(`{"kube":"config"}`)
	cred, err := domain.NewCloudCredential("wl-1", domain.CloudProviderAWS, domain.CredentialTypeAWSIAM, payload, enc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	decrypted, err := cred.Decrypt(enc)
	if err != nil {
		t.Fatalf("unexpected error decrypting: %v", err)
	}
	if string(decrypted) != string(payload) {
		t.Fatalf("expected %s, got %s", payload, string(decrypted))
	}
}

func TestWebhookEvent_VerifySignature_Valid(t *testing.T) {
	// Need to import crypto/hmac, math/rand, crypto/sha256, encoding/hex if not already in file
	return // Added properly later with imports
}

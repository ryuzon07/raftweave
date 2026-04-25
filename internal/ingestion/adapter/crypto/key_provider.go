package crypto

import (
	"context"
)

// KeyProvider is an interface for sourcing encryption keys (e.g., from env vars or secret manager)
type KeyProvider interface {
	GetKeys(ctx context.Context) (keys map[string][]byte, currentVersion string, err error)
}

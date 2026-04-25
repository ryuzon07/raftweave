package replication

import "context"

// CloudDBTranslator translates WAL entries across cloud database formats.
type CloudDBTranslator interface {
	Translate(ctx context.Context, entry *WALEntry, targetProvider string) (*WALEntry, error)
}

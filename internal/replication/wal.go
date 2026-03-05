package replication

import "context"

// WALStreamer streams Write-Ahead Log entries from the primary database.
type WALStreamer interface {
	StartStreaming(ctx context.Context, primaryEndpoint string) error
	Stop() error
	Entries() <-chan *WALEntry
}

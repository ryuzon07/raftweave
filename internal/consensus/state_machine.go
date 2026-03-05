package consensus

import "context"

// StateMachine applies committed log entries to the application state.
type StateMachine interface {
	// Apply applies a committed log entry.
	Apply(ctx context.Context, entry *LogEntry) (interface{}, error)
	// Snapshot returns a snapshot of the current state.
	Snapshot(ctx context.Context) ([]byte, error)
	// Restore restores state from a snapshot.
	Restore(ctx context.Context, data []byte) error
}

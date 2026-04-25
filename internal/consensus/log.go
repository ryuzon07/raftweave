package consensus

import "context"

// Log is the interface for the Raft log storage.
type Log interface {
	// Append appends entries to the log.
	Append(ctx context.Context, entries []LogEntry) error
	// GetEntry returns the log entry at the given index.
	GetEntry(ctx context.Context, index LogIndex) (*LogEntry, error)
	// GetEntries returns log entries from start to end (inclusive).
	GetEntries(ctx context.Context, start, end LogIndex) ([]LogEntry, error)
	// LastIndex returns the index of the last log entry.
	LastIndex() LogIndex
	// LastTerm returns the term of the last log entry.
	LastTerm() Term
	// TruncateAfter truncates all entries after the given index.
	TruncateAfter(ctx context.Context, index LogIndex) error
}

// LogEntry represents a single entry in the Raft log.
type LogEntry struct {
	Index       LogIndex
	Term        Term
	CommandType string
	Payload     []byte
}

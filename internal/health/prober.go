package health

import "context"

// Prober is the interface for health probes.
type Prober interface {
	Probe(ctx context.Context) (*ProbeResult, error)
	Name() string
}

// HTTPProber probes an HTTP endpoint.
type HTTPProber struct {
	Endpoint string
}

func (p *HTTPProber) Probe(ctx context.Context) (*ProbeResult, error) {
	return nil, nil // stub
}

func (p *HTTPProber) Name() string {
	return "http"
}

// TCPProber probes a TCP endpoint.
type TCPProber struct {
	Addr string
}

func (p *TCPProber) Probe(ctx context.Context) (*ProbeResult, error) {
	return nil, nil // stub
}

func (p *TCPProber) Name() string {
	return "tcp"
}

// DBProber probes a database connection.
type DBProber struct {
	DSN string
}

func (p *DBProber) Probe(ctx context.Context) (*ProbeResult, error) {
	return nil, nil // stub
}

func (p *DBProber) Name() string {
	return "db"
}

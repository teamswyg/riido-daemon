// Package mwsdbridge is Riido's anti-corruption layer for the
// macmini-workspace daemon.
//
// mwsd remains the local control-plane SSOT for document graph, domain DSL,
// harness history, and private repo registry. Riido consumes those contracts
// through this package instead of parsing macmini-workspace files directly.
package mwsdbridge

import (
	"time"
)

const defaultClientTimeout = 3 * time.Second

// Client reads the local mwsd Unix socket.
type Client struct {
	SocketPath string
	Timeout    time.Duration
}

// NewClient returns a Client with a conservative local timeout.
func NewClient(socketPath string) Client {
	return Client{
		SocketPath: socketPath,
		Timeout:    defaultClientTimeout,
	}
}

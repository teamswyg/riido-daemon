// Package mwsdbridge is Riido's anti-corruption layer for the
// macmini-workspace daemon.
//
// mwsd remains the local control-plane SSOT for document graph, domain DSL,
// harness history, and private repo registry. Riido consumes those contracts
// through this package instead of parsing macmini-workspace files directly.
package mwsdbridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	GraphSchemaVersion         = "mws-doc-graph.v1"
	DomainSchemaVersion        = "mws-cl-domain.v1"
	HarnessSchemaVersion       = "mws-harness-run.v1"
	ProjectsSchemaVersion      = "mws-project-registry.v1"
	OrchestrationSchemaVersion = "mws-orchestration-snapshot.v1"
)

// Client reads the local mwsd Unix socket.
type Client struct {
	SocketPath string
	Timeout    time.Duration
}

// NewClient returns a Client with a conservative local timeout.
func NewClient(socketPath string) Client {
	return Client{
		SocketPath: socketPath,
		Timeout:    3 * time.Second,
	}
}

// DefaultSocketPath returns the launchd-backed mwsd socket path.
func DefaultSocketPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "macmini-workspace", "mwsd.sock"), nil
}

// FetchSnapshot reads every mwsd contract Riido needs for its first workspace
// state projection.
func (c Client) FetchSnapshot(ctx context.Context) (Snapshot, error) {
	var snapshot Snapshot
	if err := c.Request(ctx, "status", &snapshot.Status); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, "graph", &snapshot.Graph); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, "domain", &snapshot.Domain); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, "harness", &snapshot.Harness); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, "orchestration", &snapshot.Orchestration); err != nil {
		return snapshot, err
	}
	if err := c.Request(ctx, "projects", &snapshot.Projects); err != nil {
		return snapshot, err
	}
	return snapshot, snapshot.Validate()
}

// Request sends one mwsd method request and decodes the response data.
func (c Client) Request(ctx context.Context, method string, out any) error {
	if c.SocketPath == "" {
		return errors.New("mwsd socket path is empty")
	}
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(ctx, "unix", c.SocketPath)
	if err != nil {
		return fmt.Errorf("connect mwsd socket: %w", err)
	}
	defer conn.Close()

	body, err := json.Marshal(request{Method: method})
	if err != nil {
		return fmt.Errorf("encode mwsd request: %w", err)
	}
	if _, err := conn.Write(body); err != nil {
		return fmt.Errorf("write mwsd request: %w", err)
	}
	if unix, ok := conn.(*net.UnixConn); ok {
		if err := unix.CloseWrite(); err != nil {
			return fmt.Errorf("close mwsd request stream: %w", err)
		}
	}

	responseBody, err := io.ReadAll(conn)
	if err != nil {
		return fmt.Errorf("read mwsd response: %w", err)
	}
	var env responseEnvelope
	if err := json.Unmarshal(responseBody, &env); err != nil {
		return fmt.Errorf("decode mwsd response: %w", err)
	}
	if !env.OK {
		if env.Error != "" {
			return fmt.Errorf("mwsd %s failed: %s", method, env.Error)
		}
		return fmt.Errorf("mwsd %s failed", method)
	}
	if env.Method != method {
		return fmt.Errorf("mwsd method mismatch: requested %s got %s", method, env.Method)
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("decode mwsd %s data: %w", method, err)
	}
	return nil
}

type request struct {
	Method string `json:"method"`
}

type responseEnvelope struct {
	OK     bool            `json:"ok"`
	Method string          `json:"method"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
}

// Snapshot is Riido's initial projection from macmini-workspace.
type Snapshot struct {
	Status        Status                `json:"status"`
	Graph         GraphExport           `json:"graph"`
	Domain        DomainExport          `json:"domain"`
	Harness       HarnessIndex          `json:"harness"`
	Orchestration OrchestrationSnapshot `json:"orchestration"`
	Projects      ProjectRegistry       `json:"projects"`
}

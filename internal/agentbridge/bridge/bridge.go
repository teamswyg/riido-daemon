// Package bridge owns the C4 provider-neutral high-level entry point:
// callers register one or more provider Adapters, ask for capability
// Detect, and Run a TaskRequest to receive Events + Result.
//
// It does not own concrete provider adapters, runtime scheduling, task
// persistence, EventIngestor append authority, or local API transport.
// See docs/20-domain/provider-runtime.md §7.6.
package bridge

import (
	"context"
	"errors"
	"fmt"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// Provider is the canonical adapter identifier (e.g. "claude", "codex").
type Provider string

// TaskRequest is the provider-neutral input to Run.
type TaskRequest struct {
	ID           string
	Provider     Provider
	Prompt       string
	Cwd          string
	Model        string
	SystemPrompt string
	MaxTurns     int
	// RequiredSurfaces names provider-neutral surfaces that must be
	// present before the daemon scheduler may execute the task.
	RequiredSurfaces []string `json:"required_surfaces,omitempty"`
	// AllowExperimentalRuntime opts this task into runtimes whose
	// capability snapshot requires explicit experimental use.
	AllowExperimentalRuntime bool `json:"allow_experimental_runtime,omitempty"`
	Timeout                  time.Duration
	SemanticIdle             time.Duration
	ResumeSessionID          string
	Worktree                 *assignmentcontract.AssignmentWorktree
	Env                      map[string]string
	CustomArgs               []string
	MCPConfig                []byte
	Metadata                 map[string]string
}

// RuntimeCapability pairs a provider name with its Detect snapshot.
type RuntimeCapability struct {
	Provider Provider
	Result   agentbridge.DetectResult
}

// Config carries dependencies that the caller supplies.
type Config struct {
	// Adapters MUST contain at least one provider adapter.
	Adapters []agentbridge.Adapter
	// Process is the spawn port. When nil the bridge has no way to spawn
	// — useful for tests that pre-inject a fake; production callers
	// supply a real os/exec adapter.
	Process process.Process
	// DefaultTimeout applies when TaskRequest.Timeout is zero.
	DefaultTimeout time.Duration
	// DefaultSemanticIdle applies when TaskRequest.SemanticIdle is zero.
	DefaultSemanticIdle time.Duration
	// AutoApprove is the session-level tool-approval policy. Nil → require human.
	AutoApprove agentbridge.AutoApprover
	// ToolStartGate is the session-level fail-closed policy for started tool calls.
	ToolStartGate agentbridge.ToolStartGate
}

// Client is the registry + runner.
type Client struct {
	adapters map[Provider]agentbridge.Adapter
	process  process.Process
	defaults struct {
		timeout      time.Duration
		semanticIdle time.Duration
	}
	autoApprove   agentbridge.AutoApprover
	toolStartGate agentbridge.ToolStartGate
}

// New constructs a Client. At least one adapter is required.
func New(cfg Config) (*Client, error) {
	if len(cfg.Adapters) == 0 {
		return nil, errors.New("bridge: at least one adapter is required")
	}
	c := &Client{
		adapters:      make(map[Provider]agentbridge.Adapter, len(cfg.Adapters)),
		process:       cfg.Process,
		autoApprove:   cfg.AutoApprove,
		toolStartGate: cfg.ToolStartGate,
	}
	c.defaults.timeout = cfg.DefaultTimeout
	c.defaults.semanticIdle = cfg.DefaultSemanticIdle
	for _, a := range cfg.Adapters {
		name := Provider(a.Name())
		if name == "" {
			return nil, errors.New("bridge: adapter Name() returned empty string")
		}
		if _, dup := c.adapters[name]; dup {
			return nil, fmt.Errorf("bridge: duplicate adapter %q", name)
		}
		c.adapters[name] = a
	}
	return c, nil
}

// Detect runs Detect on every registered adapter and returns the
// capability snapshots in registration order.
func (c *Client) Detect(ctx context.Context) ([]RuntimeCapability, error) {
	out := make([]RuntimeCapability, 0, len(c.adapters))
	for name, a := range c.adapters {
		res, err := a.Detect(ctx, agentbridge.DetectEnv{})
		if err != nil {
			return nil, fmt.Errorf("bridge: detect %s: %w", name, err)
		}
		out = append(out, RuntimeCapability{Provider: name, Result: res})
	}
	// Stable order: sort by provider name so callers can rely on order.
	sortByProvider(out)
	return out, nil
}

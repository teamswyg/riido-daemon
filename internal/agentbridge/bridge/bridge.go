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

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
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

// Run spawns a session for the named provider and returns a Session
// handle. The caller MUST drain Events() until it is closed; otherwise
// the session goroutine will block on send.
func (c *Client) Run(ctx context.Context, req TaskRequest) (*Session, error) {
	adapter, ok := c.adapters[req.Provider]
	if !ok {
		return nil, fmt.Errorf("bridge: unknown provider %q", req.Provider)
	}
	if c.process == nil {
		return nil, errors.New("bridge: Process port not configured")
	}

	startReq := agentbridge.StartRequest{
		TaskID:          req.ID,
		Prompt:          req.Prompt,
		Cwd:             req.Cwd,
		Model:           req.Model,
		SystemPrompt:    req.SystemPrompt,
		MaxTurns:        req.MaxTurns,
		ResumeSessionID: req.ResumeSessionID,
		Env:             req.Env,
		CustomArgs:      req.CustomArgs,
		MCPConfig:       req.MCPConfig,
		Metadata:        req.Metadata,
	}
	spawnCmd, err := adapter.BuildStart(startReq)
	if err != nil {
		return nil, fmt.Errorf("bridge: BuildStart %s: %w", req.Provider, err)
	}

	var driver agentbridge.ProtocolDriver
	if provider, ok := adapter.(agentbridge.ProtocolDriverProvider); ok {
		drv, err := provider.NewProtocolDriver(startReq)
		if err != nil {
			return nil, fmt.Errorf("bridge: NewProtocolDriver %s: %w", req.Provider, err)
		}
		driver = drv
	}

	spawnProcess := toProcessCommand(spawnCmd)
	if spawnProcess.Dir == "" {
		spawnProcess.Dir = startReq.Cwd
	}

	cfg := session.Config{
		TaskID:         req.ID,
		RuntimeID:      string(req.Provider),
		Adapter:        adapter,
		Process:        c.process,
		Spawn:          spawnProcess,
		Request:        startReq,
		HardTimeout:    firstNonZero(req.Timeout, c.defaults.timeout),
		SemanticIdle:   firstNonZero(req.SemanticIdle, c.defaults.semanticIdle),
		AutoApprove:    c.autoApprove,
		ToolStartGate:  c.toolStartGate,
		ProtocolDriver: driver,
		TempFiles:      spawnCmd.TempFiles,
	}
	inner, err := session.Start(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Session{inner: inner, droppedArgs: spawnCmd.DroppedArgs}, nil
}

// Session is the caller-facing handle for one run.
type Session struct {
	inner       *session.Session
	droppedArgs []string
}

// Events returns the run-scope event stream, closed when the session ends.
func (s *Session) Events() <-chan agentbridge.Event { return s.inner.Events() }

// Result returns the single-value terminal result channel.
func (s *Session) Result() <-chan agentbridge.Result { return s.inner.Result() }

// Cancel signals the session to terminate as ResultCancelled.
func (s *Session) Cancel(cause error) { s.inner.Cancel(cause) }

// DroppedArgs returns the custom args that BuildStart removed because
// they collided with the adapter's BlockedArgs.
func (s *Session) DroppedArgs() []string { return s.droppedArgs }

func toProcessCommand(c agentbridge.StartCommand) process.Command {
	return process.Command{
		Executable: c.Executable,
		Args:       c.Args,
		Env:        c.Env,
		Dir:        c.Dir,
	}
}

func firstNonZero(a, b time.Duration) time.Duration {
	if a > 0 {
		return a
	}
	return b
}

func sortByProvider(caps []RuntimeCapability) {
	// Simple insertion sort — order matters only for test stability.
	for i := 1; i < len(caps); i++ {
		for j := i; j > 0 && caps[j-1].Provider > caps[j].Provider; j-- {
			caps[j-1], caps[j] = caps[j], caps[j-1]
		}
	}
}

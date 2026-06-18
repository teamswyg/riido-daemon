package bridge

import (
	"errors"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// Client is the registry + runner.
type Client struct {
	adapters map[Provider]agentbridge.Adapter
	process  process.Process
	defaults clientDefaults

	autoApprove          agentbridge.AutoApprover
	toolStartGate        agentbridge.ToolStartGate
	toolApprovalGate     agentbridge.ToolApprovalGate
	toolApprovalResolver agentbridge.ToolApprovalResolver
}

type clientDefaults struct {
	timeout      time.Duration
	semanticIdle time.Duration
}

// New constructs a Client. At least one adapter is required.
func New(cfg Config) (*Client, error) {
	if len(cfg.Adapters) == 0 {
		return nil, errors.New("bridge: at least one adapter is required")
	}
	c := newClient(cfg)
	for _, adapter := range cfg.Adapters {
		if err := c.registerAdapter(adapter); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func newClient(cfg Config) *Client {
	return &Client{
		adapters:             make(map[Provider]agentbridge.Adapter, len(cfg.Adapters)),
		process:              cfg.Process,
		defaults:             clientDefaults{cfg.DefaultTimeout, cfg.DefaultSemanticIdle},
		autoApprove:          cfg.AutoApprove,
		toolStartGate:        cfg.ToolStartGate,
		toolApprovalGate:     cfg.ToolApprovalGate,
		toolApprovalResolver: cfg.ToolApprovalResolver,
	}
}

func (c *Client) registerAdapter(adapter agentbridge.Adapter) error {
	name := Provider(adapter.Name())
	if name == "" {
		return errors.New("bridge: adapter Name() returned empty string")
	}
	if _, dup := c.adapters[name]; dup {
		return fmt.Errorf("bridge: duplicate adapter %q", name)
	}
	c.adapters[name] = adapter
	return nil
}

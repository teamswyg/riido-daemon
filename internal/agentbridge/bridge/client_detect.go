package bridge

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Detect runs Detect on every registered adapter and returns the
// capability snapshots in registration order.
func (c *Client) Detect(ctx context.Context) ([]RuntimeCapability, error) {
	out := make([]RuntimeCapability, 0, len(c.adapters))
	for name, adapter := range c.adapters {
		result, err := adapter.Detect(ctx, agentbridge.DetectEnv{})
		if err != nil {
			return nil, fmt.Errorf("bridge: detect %s: %w", name, err)
		}
		out = append(out, RuntimeCapability{Provider: name, Result: result})
	}
	sortByProvider(out)
	return out, nil
}

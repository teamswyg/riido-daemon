package bridge

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type detectedRuntime struct {
	adapter agentbridge.Adapter
	detect  agentbridge.DetectResult
}

func (c *Client) resolveRuntime(ctx context.Context, req TaskRequest) (detectedRuntime, error) {
	adapter, ok := c.adapters[req.Provider]
	if !ok {
		return detectedRuntime{}, fmt.Errorf("bridge: unknown provider %q", req.Provider)
	}
	if c.process == nil {
		return detectedRuntime{}, errors.New("bridge: Process port not configured")
	}
	detect, err := adapter.Detect(ctx, agentbridge.DetectEnv{})
	if err != nil {
		return detectedRuntime{}, fmt.Errorf("bridge: detect %s: %w", req.Provider, err)
	}
	if !detect.Available {
		return detectedRuntime{}, providerUnavailableError(req.Provider, detect.Reason)
	}
	return detectedRuntime{adapter: adapter, detect: detect}, nil
}

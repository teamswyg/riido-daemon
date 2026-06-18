package runtimeactor

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func submitProtocolDriver(
	adapter agentbridge.Adapter,
	startReq agentbridge.StartRequest,
) (agentbridge.ProtocolDriver, error) {
	provider, ok := adapter.(agentbridge.ProtocolDriverProvider)
	if !ok {
		return nil, nil
	}
	driver, err := provider.NewProtocolDriver(startReq)
	if err != nil {
		return nil, fmt.Errorf("runtimeactor: NewProtocolDriver %s: %w", adapter.Name(), err)
	}
	return driver, nil
}

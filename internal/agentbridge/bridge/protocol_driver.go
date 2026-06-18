package bridge

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func newProtocolDriver(
	adapter agentbridge.Adapter,
	startReq agentbridge.StartRequest,
	providerName Provider,
) (agentbridge.ProtocolDriver, error) {
	provider, ok := adapter.(agentbridge.ProtocolDriverProvider)
	if !ok {
		return nil, nil
	}
	driver, err := provider.NewProtocolDriver(startReq)
	if err != nil {
		return nil, fmt.Errorf("bridge: NewProtocolDriver %s: %w", providerName, err)
	}
	return driver, nil
}

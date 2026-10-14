package supervisor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

type runtimePin struct {
	capabilityFingerprint string
}

func runtimePinForClaim(status runtimeactor.Status, req *bridge.TaskRequest) runtimePin {
	capView, ok := findCapability(status.Capabilities, string(req.Provider))
	if !ok {
		return runtimePin{}
	}
	return runtimePin{capabilityFingerprint: capView.CapabilityFingerprint}
}

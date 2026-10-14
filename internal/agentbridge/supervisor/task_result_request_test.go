package supervisor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func taskResultRequest() bridge.TaskRequest {
	return bridge.TaskRequest{
		ID:       "t-1",
		Provider: "fake",
		Prompt:   "hello",
		Metadata: map[string]string{
			MetadataWorkspaceID:                   "ws-1",
			MetadataAgentName:                     "Riido",
			agentbridge.MetadataTelemetryContract: agentbridge.TelemetryPlacementPrompt,
		},
	}
}

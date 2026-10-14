package runtimeactor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

const assignmentIdentityLogicalTaskID = "task-a"

func assignmentIdentityTaskRequest(id string) bridge.TaskRequest {
	return bridge.TaskRequest{
		ID:       id,
		Provider: "fake",
		Metadata: map[string]string{
			controlplane.MetadataTaskID: assignmentIdentityLogicalTaskID,
		},
	}
}

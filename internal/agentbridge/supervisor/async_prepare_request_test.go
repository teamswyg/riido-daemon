package supervisor

import (
	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func slowPrepareTaskRequest() bridge.TaskRequest {
	return bridge.TaskRequest{
		ID:                       "asn-slow-prepare",
		Provider:                 "codex",
		Prompt:                   "slow prepare",
		AllowExperimentalRuntime: true,
		Worktree: &assignmentcontract.AssignmentWorktree{
			RepositoryFullName: "teamswyg/riido-daemon",
			RepositoryURL:      "https://github.com/teamswyg/riido-daemon",
			BranchName:         "RIID-4964-agent-profile-upload",
		},
		Metadata: map[string]string{
			MetadataWorkspaceID:         "ws-1",
			MetadataRunID:               "asn-slow-prepare",
			controlplane.MetadataTaskID: "task-slow-prepare",
		},
	}
}

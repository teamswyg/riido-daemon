package supervisor

import (
	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func prepareCancellationRequest() bridge.TaskRequest {
	return bridge.TaskRequest{
		ID:                       "t-cancel-prepare",
		Provider:                 "codex",
		Prompt:                   "x",
		AllowExperimentalRuntime: true,
		Worktree: &assignmentcontract.AssignmentWorktree{
			RepositoryFullName: "teamswyg/riido-daemon",
			RepositoryURL:      "https://github.com/teamswyg/riido-daemon",
			BranchName:         "RIID-4964-agent-profile-upload",
		},
		Metadata: map[string]string{
			MetadataWorkspaceID:         "ws-1",
			MetadataRunID:               "run-cancel-prepare",
			controlplane.MetadataTaskID: "task-cancel-prepare",
		},
	}
}

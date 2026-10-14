package saasplane

import assignmentcontract "github.com/teamswyg/riido-contracts/assignment"

func saasClaimReportAssignment() assignmentcontract.Assignment {
	return assignmentcontract.Assignment{
		ID:               "asn-1",
		TaskID:           "task-a",
		ComponentID:      "component-1",
		AgentID:          "jykim1",
		RuntimeProvider:  "codex",
		Prompt:           "golang hello world quickly",
		AgentInstruction: "write concise Korean progress updates",
		ResumeSessionID:  "sess-prev",
		Worktree: &assignmentcontract.AssignmentWorktree{
			RepositoryFullName: "teamswyg/riido-daemon",
			RepositoryURL:      "https://github.com/teamswyg/riido-daemon",
			BranchName:         "RIID-4964-agent-profile-upload",
			Source:             "connected_pull_request",
		},
		State:      assignmentcontract.AssignmentQueued,
		LeaseToken: "lease-1",
	}
}

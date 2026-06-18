package main

import (
	"time"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func validateTestProviderCandidates() []taskdb.ProviderCandidate {
	return []taskdb.ProviderCandidate{
		{
			ID:               "codex",
			SourceWorkflow:   "provider-selection",
			Available:        true,
			ApprovalRequired: true,
		},
	}
}

func newValidateTestTask(state task.TaskState, now time.Time) taskdb.TaskRecord {
	return taskdb.TaskRecord{
		ID:                     validateTestTaskID,
		ProjectID:              "macmini-workspace",
		State:                  state,
		SourceDocumentID:       "mws.test",
		SourceDocumentPath:     "docs/TEST.md",
		Title:                  "Validation FSM test",
		Owner:                  "human",
		SourceStatus:           "in-progress",
		RecommendedProvider:    "codex",
		RecommendedDecisionLLM: "codex",
		RequiresHumanApproval:  true,
		HarnessNextDirection:   "top-down",
		CreatedAt:              now.Format(time.RFC3339Nano),
		UpdatedAt:              now.Format(time.RFC3339Nano),
	}
}

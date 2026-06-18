package taskvalidation

import (
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

const validationFixtureTime = "2026-05-20T08:00:00Z"

func validationProviderCandidates() []taskdb.ProviderCandidate {
	return []taskdb.ProviderCandidate{
		{ID: "codex", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true},
	}
}

func validationTaskRecord(state task.TaskState) taskdb.TaskRecord {
	return taskdb.TaskRecord{
		ID:                     "task:test",
		ProjectID:              "macmini-workspace",
		State:                  state,
		SourceDocumentID:       "mws.test",
		SourceDocumentPath:     "docs/TEST.md",
		Title:                  "test",
		RecommendedProvider:    "codex",
		RecommendedDecisionLLM: "codex",
		RequiresHumanApproval:  true,
		CreatedAt:              validationFixtureTime,
		UpdatedAt:              validationFixtureTime,
	}
}

func validationCreatedTransition() taskdb.TaskTransitionRecord {
	return taskdb.TaskTransitionRecord{
		ID:         "transition:task:test:created",
		TaskID:     "task:test",
		ToState:    task.StateCreated,
		EventType:  ir.EventTaskCreated,
		Actor:      "riido",
		Source:     "test",
		RecordedAt: validationFixtureTime,
	}
}

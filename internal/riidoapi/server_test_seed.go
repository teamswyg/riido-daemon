package riidoapi

import (
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func seedTaskDB(state task.TaskState) taskdb.TaskDB {
	db := taskdb.EmptyTaskDB()
	db.UpdatedAt = "2026-05-20T08:00:00Z"
	db.RecommendedProvider = "codex"
	db.RecommendedDecisionLLM = "codex"
	db.DecisionGate = "human-approval-required"
	db.ProviderCandidates = []taskdb.ProviderCandidate{
		{ID: "codex", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true},
	}
	db.Tasks = []taskdb.TaskRecord{seedTaskRecord(state)}
	db.Transitions = []taskdb.TaskTransitionRecord{seedTransition()}
	return db
}

func seedTaskRecord(state task.TaskState) taskdb.TaskRecord {
	return taskdb.TaskRecord{
		ID:                     "task:test",
		ProjectID:              "macmini-workspace",
		State:                  state,
		SourceDocumentID:       "mws.test",
		SourceDocumentPath:     "docs/TEST.md",
		Title:                  "테스트",
		Owner:                  "local",
		SourceStatus:           "seed",
		RecommendedProvider:    "codex",
		RecommendedDecisionLLM: "codex",
		RequiresHumanApproval:  true,
		HarnessNextDirection:   "top-down",
		CreatedAt:              "2026-05-20T08:00:00Z",
		UpdatedAt:              "2026-05-20T08:00:00Z",
		TransitionCount:        1,
	}
}

func seedTransition() taskdb.TaskTransitionRecord {
	return taskdb.TaskTransitionRecord{
		ID:         "transition:task:test:created",
		TaskID:     "task:test",
		ToState:    task.StateCreated,
		EventType:  ir.EventTaskCreated,
		Actor:      "riido",
		Source:     "test",
		RecordedAt: "2026-05-20T08:00:00Z",
	}
}

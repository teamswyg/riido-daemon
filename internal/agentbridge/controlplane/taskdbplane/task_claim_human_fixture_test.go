package taskdbplane

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func approvedHumanClaimDB(t *testing.T) taskdb.TaskDB {
	t.Helper()
	db := humanApprovalGateDB(task.StateCreated)
	db, _, _, err := taskdb.ApplyGuardedTaskTransition(
		db,
		approvedHumanQueueTransition(),
		time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("queue transition: %v", err)
	}
	return db
}

func humanApprovalGateDB(state task.TaskState) taskdb.TaskDB {
	return taskdb.TaskDB{
		SchemaVersion:          taskdb.TaskDBSchemaVersion,
		DecisionGate:           "human-approval-required",
		RecommendedProvider:    "codex",
		RecommendedDecisionLLM: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
		},
		Tasks: []taskdb.TaskRecord{{
			ID:                     "task-human",
			ProjectID:              "project-1",
			State:                  state,
			Title:                  "approved task",
			RecommendedProvider:    "codex",
			RecommendedDecisionLLM: "codex",
			RequiresHumanApproval:  true,
		}},
	}
}

func approvedHumanQueueTransition() taskdb.TaskTransitionInput {
	return taskdb.TaskTransitionInput{
		TaskID:  "task-human",
		ToState: task.StateQueued,
		Event:   ir.EventTaskQueued,
		Actor:   "human",
		Source:  "test",
		Reason:  "approved for run",
		Guard: taskdb.TaskMutationGuardInput{
			CommandID:   "command:test:queue",
			Provider:    "codex",
			DecisionLLM: "codex",
			ApprovalID:  "approval:human:1",
		},
	}
}

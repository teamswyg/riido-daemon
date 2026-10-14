package taskvalidation

import (
	"time"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func validationTaskDB(state task.TaskState) taskdb.TaskDB {
	db := taskdb.EmptyTaskDB()
	db.UpdatedAt = validationFixtureTime
	db.RecommendedProvider = "codex"
	db.RecommendedDecisionLLM = "codex"
	db.DecisionGate = "human-approval-required"
	db.ProviderCandidates = validationProviderCandidates()
	db.Tasks = []taskdb.TaskRecord{validationTaskRecord(state)}
	db.Transitions = []taskdb.TaskTransitionRecord{validationCreatedTransition()}
	return db
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 20, 8, 0, 0, 0, time.UTC)
}

package taskdb

import (
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

func TestTaskDBInputErrorIsClassified(t *testing.T) {
	_, _, _, err := ApplyGuardedTaskTransition(TaskDB{}, TaskTransitionInput{}, time.Now())
	if err == nil {
		t.Fatal("expected input error")
	}
	if !errors.Is(err, ErrTaskDBInput) {
		t.Fatalf("errors.Is(err, ErrTaskDBInput) = false for %v", err)
	}
}

func TestTaskDBGuardErrorIsClassified(t *testing.T) {
	db := EmptyTaskDB()
	db.Tasks = []TaskRecord{{
		ID:                    "task-1",
		State:                 task.StateQueued,
		RequiresHumanApproval: true,
		RecommendedProvider:   "codex",
	}}
	db.ProviderCandidates = []ProviderCandidate{{ID: "codex", Available: true}}

	_, _, _, err := ApplyGuardedTaskTransition(db, TaskTransitionInput{
		TaskID:  "task-1",
		ToState: task.StateClaimed,
		Event:   ir.EventTaskClaimed,
	}, time.Now())
	if err == nil {
		t.Fatal("expected guard error")
	}
	if !errors.Is(err, ErrTaskDBGuard) {
		t.Fatalf("errors.Is(err, ErrTaskDBGuard) = false for %v", err)
	}
}

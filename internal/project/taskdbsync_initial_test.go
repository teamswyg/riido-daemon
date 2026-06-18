package project

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestSyncTaskDBFromStateCreatesInitialTransitions(t *testing.T) {
	state := sampleTaskDBState(t)

	db := SyncTaskDBFromState(taskdb.EmptyTaskDB(), state, taskDBSyncTime(10))

	assertInitialTaskDBShape(t, db)
	assertInitialTaskDBRecommendations(t, db)
	assertInitialTaskDBFirstTask(t, db)
}

func assertInitialTaskDBShape(t *testing.T, db taskdb.TaskDB) {
	t.Helper()

	if db.SchemaVersion != taskdb.TaskDBSchemaVersion {
		t.Fatalf("unexpected task DB schema: %s", db.SchemaVersion)
	}
	if len(db.Tasks) != 2 {
		t.Fatalf("unexpected task count: %d", len(db.Tasks))
	}
	if len(db.Transitions) != 2 {
		t.Fatalf("unexpected transition count: %d", len(db.Transitions))
	}
	if len(db.Evidence) != 0 {
		t.Fatalf("unexpected evidence count: %d", len(db.Evidence))
	}
	if len(db.CommandReceipts) != 0 {
		t.Fatalf("sync should not create command receipts: %d", len(db.CommandReceipts))
	}
}

func assertInitialTaskDBRecommendations(t *testing.T, db taskdb.TaskDB) {
	t.Helper()

	if db.RecommendedProvider != "codex" {
		t.Fatalf("unexpected DB recommended provider: %s", db.RecommendedProvider)
	}
	if db.RecommendedDecisionLLM != "codex" {
		t.Fatalf("unexpected DB decision LLM: %s", db.RecommendedDecisionLLM)
	}
	if db.DecisionGate != "human-approval-required" {
		t.Fatalf("unexpected DB decision gate: %s", db.DecisionGate)
	}
	if len(db.ProviderCandidates) != 3 {
		t.Fatalf("unexpected DB provider candidate count: %d", len(db.ProviderCandidates))
	}
}

func assertInitialTaskDBFirstTask(t *testing.T, db taskdb.TaskDB) {
	t.Helper()

	if db.Tasks[0].ID != "task:mws.goal" {
		t.Fatalf("tasks should be sorted by id: %#v", db.Tasks)
	}
	if db.Tasks[0].State != task.StateCreated {
		t.Fatalf("unexpected initial state: %s", db.Tasks[0].State)
	}
	if db.Transitions[0].EventType != ir.EventTaskCreated {
		t.Fatalf("unexpected initial event: %s", db.Transitions[0].EventType)
	}
	if db.Tasks[0].TransitionCount != 1 {
		t.Fatalf("unexpected transition count on first task: %d", db.Tasks[0].TransitionCount)
	}
}

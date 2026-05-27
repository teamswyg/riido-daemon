package project

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestSyncTaskDBFromStateCreatesInitialTransitions(t *testing.T) {
	state := sampleState(t)
	now := time.Date(2026, 5, 20, 8, 10, 0, 0, time.UTC)

	db := SyncTaskDBFromState(taskdb.EmptyTaskDB(), state, now)
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
	if db.Tasks[0].RecommendedProvider != "codex" {
		t.Fatalf("unexpected task recommended provider: %s", db.Tasks[0].RecommendedProvider)
	}
	if db.Tasks[0].RecommendedDecisionLLM != "codex" {
		t.Fatalf("unexpected task decision LLM: %s", db.Tasks[0].RecommendedDecisionLLM)
	}
	if !db.Tasks[0].RequiresHumanApproval {
		t.Fatalf("task should require human approval")
	}
}

func TestSyncTaskDBFromStateDoesNotDuplicateInitialTransitions(t *testing.T) {
	state := sampleState(t)
	first := SyncTaskDBFromState(taskdb.EmptyTaskDB(), state, time.Date(2026, 5, 20, 8, 10, 0, 0, time.UTC))
	second := SyncTaskDBFromState(first, state, time.Date(2026, 5, 20, 8, 11, 0, 0, time.UTC))

	if len(second.Transitions) != len(first.Transitions) {
		t.Fatalf("sync should not duplicate initial transitions: first=%d second=%d", len(first.Transitions), len(second.Transitions))
	}
	record := findTaskRecord(t, second, "task:mws.goal")
	if record.TransitionCount != 1 {
		t.Fatalf("unexpected transition count after second sync: %d", record.TransitionCount)
	}
}

func TestSyncTaskDBFromStateUpdatesOrchestrationMetadata(t *testing.T) {
	state := sampleState(t)
	first := SyncTaskDBFromState(taskdb.EmptyTaskDB(), state, time.Date(2026, 5, 20, 8, 10, 0, 0, time.UTC))
	state.RecommendedProvider = "claude"
	state.RecommendedDecisionLLM = "codex"
	state.ProviderCandidates = []ProviderCandidate{
		{ID: "claude", SourceWorkflow: "provider-selection", Available: true, ApprovalRequired: true},
	}
	for index := range state.Tasks {
		state.Tasks[index].RecommendedProvider = "claude"
		state.Tasks[index].RecommendedDecisionLLM = "codex"
	}

	second := SyncTaskDBFromState(first, state, time.Date(2026, 5, 20, 8, 12, 0, 0, time.UTC))
	if second.RecommendedProvider != "claude" {
		t.Fatalf("sync should update DB recommended provider: %s", second.RecommendedProvider)
	}
	if len(second.ProviderCandidates) != 1 || second.ProviderCandidates[0].ID != "claude" {
		t.Fatalf("sync should update provider candidates: %#v", second.ProviderCandidates)
	}
	record := findTaskRecord(t, second, "task:mws.goal")
	if record.RecommendedProvider != "claude" {
		t.Fatalf("sync should update task recommended provider: %s", record.RecommendedProvider)
	}
	if len(second.Transitions) != len(first.Transitions) {
		t.Fatalf("metadata sync should not append transitions: first=%d second=%d", len(first.Transitions), len(second.Transitions))
	}
}

func sampleState(t *testing.T) StateFile {
	t.Helper()
	projection, err := FromMwsdSnapshot(sampleSnapshot())
	if err != nil {
		t.Fatalf("FromMwsdSnapshot returned error: %v", err)
	}
	return StateFromProjection(projection)
}

func findTaskRecord(t *testing.T, db taskdb.TaskDB, taskID string) taskdb.TaskRecord {
	t.Helper()
	for _, record := range db.Tasks {
		if record.ID == taskID {
			return record
		}
	}
	t.Fatalf("task %s not found", taskID)
	return taskdb.TaskRecord{}
}

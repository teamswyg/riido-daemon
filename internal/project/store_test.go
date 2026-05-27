package project

import (
	"path/filepath"
	"testing"
)

func TestStateFromProjection(t *testing.T) {
	projection, err := FromMwsdSnapshot(sampleSnapshot())
	if err != nil {
		t.Fatalf("FromMwsdSnapshot returned error: %v", err)
	}

	state := StateFromProjection(projection)
	if state.SchemaVersion != StateSchemaVersion {
		t.Fatalf("unexpected state schema: %s", state.SchemaVersion)
	}
	if state.ProjectionVersion != "riido-workspace-projection.v1" {
		t.Fatalf("unexpected projection schema: %s", state.ProjectionVersion)
	}
	if state.RecommendedProvider != "codex" {
		t.Fatalf("unexpected recommended provider: %s", state.RecommendedProvider)
	}
	if state.RecommendedDecisionLLM != "codex" {
		t.Fatalf("unexpected recommended decision LLM: %s", state.RecommendedDecisionLLM)
	}
	if state.DecisionGate != "human-approval-required" {
		t.Fatalf("unexpected decision gate: %s", state.DecisionGate)
	}
	if len(state.ProviderCandidates) != 3 {
		t.Fatalf("unexpected provider candidate count: %d", len(state.ProviderCandidates))
	}
	if len(state.Projects) != 3 {
		t.Fatalf("unexpected project count: %d", len(state.Projects))
	}
	if len(state.Tasks) != 2 {
		t.Fatalf("unexpected task count: %d", len(state.Tasks))
	}
	if state.Tasks[0].ID != "task:mws.goal" {
		t.Fatalf("unexpected first task: %#v", state.Tasks[0])
	}
	if state.Tasks[0].State != "Created" {
		t.Fatalf("unexpected initial task state: %s", state.Tasks[0].State)
	}
	if state.Tasks[0].RecommendedProvider != "codex" {
		t.Fatalf("unexpected task recommended provider: %s", state.Tasks[0].RecommendedProvider)
	}
	if state.Tasks[0].RecommendedDecisionLLM != "codex" {
		t.Fatalf("unexpected task recommended decision LLM: %s", state.Tasks[0].RecommendedDecisionLLM)
	}
	if !state.Tasks[0].RequiresHumanApproval {
		t.Fatalf("task should carry human approval gate")
	}
}

func TestSaveAndLoadState(t *testing.T) {
	projection, err := FromMwsdSnapshot(sampleSnapshot())
	if err != nil {
		t.Fatalf("FromMwsdSnapshot returned error: %v", err)
	}
	path := filepath.Join(t.TempDir(), "workspace-state.json")

	if err := SaveState(path, StateFromProjection(projection)); err != nil {
		t.Fatalf("SaveState returned error: %v", err)
	}
	loaded, err := LoadState(path)
	if err != nil {
		t.Fatalf("LoadState returned error: %v", err)
	}
	if loaded.SchemaVersion != StateSchemaVersion {
		t.Fatalf("unexpected loaded schema: %s", loaded.SchemaVersion)
	}
	if loaded.Tasks[1].ID != "task:mws.roadmap" {
		t.Fatalf("unexpected loaded task: %#v", loaded.Tasks[1])
	}
}

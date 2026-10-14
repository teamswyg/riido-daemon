package project

import "testing"

func TestStateFromProjection(t *testing.T) {
	projection, err := FromMwsdSnapshot(sampleSnapshot())
	if err != nil {
		t.Fatalf("FromMwsdSnapshot returned error: %v", err)
	}

	state := StateFromProjection(projection)
	assertStateProjectionMetadata(t, state)
	assertStateProjectionCollections(t, state)
	assertStateProjectionFirstTask(t, state)
}

func assertStateProjectionMetadata(t *testing.T, state StateFile) {
	t.Helper()
	if state.SchemaVersion != StateSchemaVersion ||
		state.ProjectionVersion != "riido-workspace-projection.v1" ||
		state.RecommendedProvider != "codex" ||
		state.RecommendedDecisionLLM != "codex" ||
		state.DecisionGate != "human-approval-required" {
		t.Fatalf("unexpected state metadata: %+v", state)
	}
}

func assertStateProjectionCollections(t *testing.T, state StateFile) {
	t.Helper()
	if len(state.ProviderCandidates) != 3 || len(state.Projects) != 3 || len(state.Tasks) != 2 {
		t.Fatalf("unexpected state collection sizes: %+v", state)
	}
}

func assertStateProjectionFirstTask(t *testing.T, state StateFile) {
	t.Helper()
	task := state.Tasks[0]
	if task.ID != "task:mws.goal" ||
		task.State != "Created" ||
		task.RecommendedProvider != "codex" ||
		task.RecommendedDecisionLLM != "codex" ||
		!task.RequiresHumanApproval {
		t.Fatalf("unexpected first task: %#v", task)
	}
}

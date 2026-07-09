package main

import "testing"

func TestTaskSummarySkipAndConfiguredTeamEdges(t *testing.T) {
	agents := summarizeAssignableAgents(map[string]any{"agents": []any{"a", "b", "c"}})
	if agents["agents_count"] != 3 {
		t.Fatalf("agents summary=%#v", agents)
	}
	in := scenario{ID: "contract.task.example", Observed: map[string]any{"before": true}}
	skipped := skippedTaskScenario(in, "missing task config")
	if skipped.Status != statusSkipped || skipped.Repair == nil {
		t.Fatalf("skipped scenario=%#v", skipped)
	}
	teamID := "team-configured"
	workspaceID := "workspace-1"
	got, sc := taskFixtureTeamID(apiClient{}, config{teamID: &teamID, workspaceID: &workspaceID})
	if got != teamID || sc.Status != statusPassed {
		t.Fatalf("team=%q scenario=%#v", got, sc)
	}
	if sc.Observed["team_id_source"] != "configured" {
		t.Fatalf("team source not recorded: %#v", sc)
	}
}

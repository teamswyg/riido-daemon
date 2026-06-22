package main

import "testing"

func TestSummarizeFailsWhenAnyScenarioFails(t *testing.T) {
	got := summarize([]scenario{
		{ID: "a", Status: statusPassed},
		{ID: "b", Status: statusFailed},
	})
	if got != statusFailed {
		t.Fatalf("status=%q", got)
	}
}

func TestWorkspaceRoutesRequireWorkspaceID(t *testing.T) {
	got := workspaceRouteScenarios("http://127.0.0.1:1", "")
	if len(got) != 3 || got[0].Status != statusSkipped {
		t.Fatalf("routes=%+v", got)
	}
	if got[0].Repair == nil || got[0].Repair.Class != "workspace_id_required" {
		t.Fatalf("repair=%+v", got[0].Repair)
	}
}

func TestMissingRouteDetection(t *testing.T) {
	if !isMissingRoute(routeProbe{Body: "404 찾을 수 없는 페이지"}) {
		t.Fatal("missing route was not detected")
	}
}

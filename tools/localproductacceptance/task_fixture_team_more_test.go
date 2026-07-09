package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskFixtureTeamIDUsesConfiguredTeamWithoutAPI(t *testing.T) {
	teamID := "team-configured"
	cfg := config{teamID: &teamID}
	got, sc := taskFixtureTeamID(apiClient{}, cfg)
	if got != teamID || sc.Status != statusPassed {
		t.Fatalf("team=%q scenario=%+v", got, sc)
	}
	if sc.Observed["team_id_source"] != "configured" {
		t.Fatalf("scenario observed=%+v", sc.Observed)
	}
}

func TestTaskFixtureTeamIDFailsWhenWorkspaceTeamsHaveNoID(t *testing.T) {
	var requested string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = r.Method + " " + r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]any{"teams": []any{map[string]any{"name": "No ID"}}})
	}))
	defer server.Close()

	workspaceID, emptyTeam := "workspace-a", ""
	cfg := config{workspaceID: &workspaceID, teamID: &emptyTeam}
	got, sc := taskFixtureTeamID(newAPIClient(server.URL, "token"), cfg)
	if got != "" || sc.Status != statusFailed {
		t.Fatalf("team=%q scenario=%+v", got, sc)
	}
	if requested != "GET /workspaces/workspace-a/teams" {
		t.Fatalf("requested=%q", requested)
	}
	if sc.Repair == nil || sc.Repair.Class != "task_fixture_create_failed" {
		t.Fatalf("repair=%+v", sc.Repair)
	}
}

func TestTaskFixtureRepairCarriesSummary(t *testing.T) {
	rep := taskFixtureRepair("team lookup failed")
	if rep.Class != "task_fixture_create_failed" ||
		rep.Owner != "riido-api-server/local-qa" ||
		rep.Summary != "team lookup failed" {
		t.Fatalf("repair=%+v", rep)
	}
}

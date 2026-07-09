package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMaybeCreateTaskFixtureCreatesConfiguredTeamFixture(t *testing.T) {
	var seen string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Method + " " + r.URL.Path
		if seen != "POST /teams/team-a/components" {
			t.Fatalf("unexpected request %s", seen)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "task-a"})
	}))
	defer server.Close()

	cfg := fixtureConfig(server.URL)
	teamID, workspaceID, enabled := "team-a", "workspace-a", true
	cfg.teamID = &teamID
	cfg.workspaceID = &workspaceID
	cfg.taskFixture = &enabled

	fixture := maybeCreateTaskFixture(cfg, "generated")
	if !fixture.Created() || fixture.TaskID != "task-a" || fixture.TeamID != "team-a" {
		t.Fatalf("fixture=%+v", fixture)
	}
	if fixture.Team.Status != statusPassed || fixture.Create.Status != statusPassed {
		t.Fatalf("team=%+v create=%+v", fixture.Team, fixture.Create)
	}
	if seen == "" {
		t.Fatal("expected create request")
	}
}

func TestMaybeCreateTaskFixtureNoopsWhenDisabled(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		calls++
	}))
	defer server.Close()

	cfg := fixtureConfig(server.URL)
	teamID, workspaceID := "team-a", "workspace-a"
	cfg.teamID = &teamID
	cfg.workspaceID = &workspaceID

	for name, tc := range map[string]struct {
		source      string
		runMutation bool
		taskFixture bool
	}{
		"source":     {"manual", true, true},
		"mutation":   {"generated", false, true},
		"fixtureOff": {"generated", true, false},
	} {
		t.Run(name, func(t *testing.T) {
			cfg.runMutations = &tc.runMutation
			cfg.taskFixture = &tc.taskFixture
			if fixture := maybeCreateTaskFixture(cfg, tc.source); fixture.Created() {
				t.Fatalf("fixture=%+v", fixture)
			}
		})
	}
	if calls != 0 {
		t.Fatalf("unexpected API calls=%d", calls)
	}
}

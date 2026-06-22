package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAndCleanupTaskFixtureUseRiidoTaskEndpoints(t *testing.T) {
	var created, deleted string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/teams/team-a/components":
			created = r.URL.Path
			_ = json.NewEncoder(w).Encode(map[string]any{"componentId": "task-a"})
		case r.Method == http.MethodDelete && r.URL.Path == "/teams/team-a/components/task-a":
			deleted = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	cfg := fixtureConfig(server.URL)
	fixture := createTaskFixture(newAPIClient(server.URL, "token"), "team-a")
	if !fixture.Created() || created == "" {
		t.Fatalf("fixture=%+v created=%q", fixture, created)
	}
	cleanup := cleanupTaskFixture(cfg, fixture)
	if cleanup.Status != statusPassed || deleted == "" {
		t.Fatalf("cleanup=%+v deleted=%q", cleanup, deleted)
	}
}

func fixtureConfig(host string) config {
	token, enabled := "token", true
	return config{riidoAPIHost: &host, apiToken: &token, runMutations: &enabled}
}

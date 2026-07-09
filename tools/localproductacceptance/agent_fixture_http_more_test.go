package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateAndCleanupAgentFixturePair(t *testing.T) {
	created := 0
	var deleted []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/base/agents":
			created++
			fmt.Fprintf(w, `{"agent":{"agent_id":"agent_%d",`+
				`"runtime_id":"rt_%d","runtime_kind":"codex","visibility":"private"}}`,
				created, created)
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/base/agents/"):
			deleted = append(deleted, strings.TrimPrefix(r.URL.Path, "/base/agents/"))
			fmt.Fprint(w, `{"agent_id":"deleted","queued_tasks_unassigned":1,`+
				`"running_tasks_force_stopped":0}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	pair := [2]preparedRuntime{
		{RuntimeID: "rt_1", Kind: "codex"},
		{RuntimeID: "rt_2", Kind: "codex"},
	}
	fixture := createAgentFixturePair(newAPIClient(server.URL, "tok"), "/base", pair)
	if len(fixture.Candidates) != 2 || len(fixture.CreatedIDs) != 2 {
		t.Fatalf("unexpected fixture: %#v", fixture)
	}
	if fixture.Candidates[0].RuntimeID != "rt_1" || fixture.Candidates[1].RuntimeID != "rt_2" {
		t.Fatalf("runtime ids should come from prepared pair: %#v", fixture.Candidates)
	}
	cleanup := cleanupAgentFixtures(newAPIClient(server.URL, "tok"), "/base", fixture)
	if len(cleanup) != 2 || len(deleted) != 2 {
		t.Fatalf("unexpected cleanup: %#v deleted=%#v", cleanup, deleted)
	}
}

func TestMaybeCreateAgentFixturesGuardsMutationFlagsAndRuntimePair(t *testing.T) {
	off := false
	on := true
	cfg := config{prepareDaemon: &off, runMutations: &on}
	fixture := maybeCreateAgentFixtures(cfg, apiClient{}, "/base", saasPrepareResult{})
	if len(fixture.Scenarios) != 0 || len(fixture.Candidates) != 0 {
		t.Fatalf("fixtures should be disabled by prepare flag: %#v", fixture)
	}
	cfg.prepareDaemon = &on
	cfg.runMutations = &on
	fixture = maybeCreateAgentFixtures(cfg, apiClient{}, "/base", saasPrepareResult{})
	if len(fixture.Scenarios) != 1 || fixture.Scenarios[0].Status != statusSkipped {
		t.Fatalf("missing runtime pair should be skipped: %#v", fixture)
	}
}

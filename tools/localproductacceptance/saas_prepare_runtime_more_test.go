package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWaitForPreparedSaaSRuntimesAppendsReadySnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/client/workspaces/workspace-a/ai-agent/devices" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"devices":[
{"device_id":"dev-a","runtimes":[{"runtime_id":"rt-a","kind":"codex","provider_version":"v1","availability":"online","detection_state":"detected"}]},
{"device_id":"dev-b","runtimes":[{"runtime_id":"rt-b","kind":"codex","provider_version":"v1","availability":"online","detection_state":"detected"}]}
]}`)
	}))
	defer server.Close()

	cfg := contractAPIFlowConfig(server.URL)
	prep := saasPrepareResult{DeviceIDs: []string{"dev-a", "dev-b"}}
	got := waitForPreparedSaaSRuntimes(cfg, newAPIClient(server.URL, "tok"), prep)
	if len(got.Scenarios) != 1 || got.Scenarios[0].Status != statusPassed {
		t.Fatalf("prep=%+v", got)
	}
	if len(got.Runtimes) != 2 || got.Runtimes[0].ProviderVersion != "v1" {
		t.Fatalf("runtimes=%+v", got.Runtimes)
	}
}

func TestMaybeCreateAgentFixturesCreatesPreparedRuntimePair(t *testing.T) {
	created := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/base/agents" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		created++
		fmt.Fprintf(w, `{"agent":{"agent_id":"agent-%d","runtime_id":"rt-%d","runtime_kind":"codex"}}`,
			created, created)
	}))
	defer server.Close()

	on := true
	cfg := config{prepareDaemon: &on, runMutations: &on}
	prep := saasPrepareResult{Runtimes: []preparedRuntime{
		{RuntimeID: "rt-1", Kind: "codex"},
		{RuntimeID: "rt-2", Kind: "codex"},
	}}
	fixture := maybeCreateAgentFixtures(cfg, newAPIClient(server.URL, "tok"), "/base", prep)
	if len(fixture.Candidates) != 2 || len(fixture.Scenarios) != 3 {
		t.Fatalf("fixture=%+v", fixture)
	}
	if fixture.Scenarios[2].ID != "contract.task.frontend_identity_contract" {
		t.Fatalf("identity scenario missing: %+v", fixture.Scenarios)
	}
}

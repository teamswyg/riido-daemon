package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContractAPIScenariosRunsReadOnlySmokeWhenConfigured(t *testing.T) {
	seen := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen[r.Method+" "+r.URL.Path]++
		switch {
		case r.Method == http.MethodGet && suffix(r.URL.Path, "/bootstrap"):
			_ = json.NewEncoder(w).Encode(map[string]any{"workspace_id": "workspace-a"})
		case r.Method == http.MethodGet && suffix(r.URL.Path, "/devices"):
			_ = json.NewEncoder(w).Encode(map[string]any{"devices": []any{}})
		case r.Method == http.MethodPost && suffix(r.URL.Path, "/profile-thumbnails/uploads"):
			_ = json.NewEncoder(w).Encode(map[string]any{"upload_url": "https://s3.example/upload"})
		case r.Method == http.MethodGet && suffix(r.URL.Path, "/tasks/assigned-agent-profiles"):
			_ = json.NewEncoder(w).Encode(map[string]any{"assigned_agent_profiles": map[string]any{}})
		case r.Method == http.MethodGet && suffix(r.URL.Path, "/tasks/task-a/assignable-agents"):
			_ = json.NewEncoder(w).Encode(map[string]any{"agents": []any{}})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	cfg := contractAPIFlowConfig(server.URL)
	scenarios := contractAPIScenarios(cfg)
	if len(scenarios) != 9 {
		t.Fatalf("scenarios=%#v", scenarios)
	}
	for _, sc := range scenarios[:5] {
		if sc.Status != statusPassed {
			t.Fatalf("read scenario failed: %#v", sc)
		}
	}
	for _, want := range []string{
		"GET /v2/client/workspaces/workspace-a/ai-agent/bootstrap",
		"GET /v2/client/workspaces/workspace-a/ai-agent/devices",
		"POST /v2/client/workspaces/workspace-a/ai-agent/profile-thumbnails/uploads",
		"GET /v2/client/workspaces/workspace-a/ai-agent/tasks/assigned-agent-profiles",
		"GET /v2/client/workspaces/workspace-a/ai-agent/tasks/task-a/assignable-agents",
	} {
		if seen[want] != 1 {
			t.Fatalf("missing request %s in %#v", want, seen)
		}
	}
}

func contractAPIFlowConfig(host string) config {
	workspaceID, taskID, token := "workspace-a", "task-a", "token"
	empty := ""
	disabled := false
	slots := 0
	return config{
		workspaceID: &workspaceID, taskID: &taskID, apiToken: &token,
		agentHost: &host, riidoAPIHost: &host, teamID: &empty,
		firstAgentID: &empty, secondAgentID: &empty, commentBody: &empty,
		runMutations: &disabled, taskFixture: &disabled,
		prepareDaemon: &disabled, daemonBinary: &empty, daemonSlots: &slots,
	}
}

func suffix(path, tail string) bool {
	return len(path) >= len(tail) && path[len(path)-len(tail):] == tail
}

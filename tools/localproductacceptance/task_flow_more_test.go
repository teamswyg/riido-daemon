package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskFlowScenariosSkipMutationsAfterAssignableAgents(t *testing.T) {
	var requested string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = r.Method + " " + r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]any{"agents": []any{
			map[string]any{"agent_id": "agent-a", "runtime_kind": "codex"},
			map[string]any{"agent_id": "agent-b", "runtime_kind": "codex"},
		}})
	}))
	defer server.Close()

	cfg := taskFlowTestConfig(server.URL, false)
	scenarios := taskFlowScenarios(newAPIClient(server.URL, "token"), cfg, nil, agentFixture{})
	if requested != "GET /v2/client/workspaces/workspace-a/ai-agent/tasks/task-a/assignable-agents" {
		t.Fatalf("requested=%q", requested)
	}
	if len(scenarios) != 5 || scenarios[0].ID != "contract.task.assignable_agents" {
		t.Fatalf("scenarios=%+v", scenarios)
	}
	for _, sc := range scenarios[1:] {
		if sc.Status != statusSkipped || sc.Repair == nil {
			t.Fatalf("expected mutation skip repair: %+v", sc)
		}
	}
}

func TestTaskFlowHelpersForGeneratedFailureAndSkipTail(t *testing.T) {
	assignable := scenario{ID: "assignable", Status: statusFailed}
	if !shouldSkipGeneratedTaskFlow(assignable, "generated") {
		t.Fatal("generated failed assignable response should skip flow")
	}
	tail := taskSkipped(true, "need task")
	if len(tail) != 5 || tail[0].ID != "contract.task.assignable_agents" {
		t.Fatalf("tail=%+v", tail)
	}
}

func taskFlowTestConfig(host string, mutate bool) config {
	workspaceID := "workspace-a"
	taskID := "task-a"
	teamID := ""
	token := "token"
	fixture := false
	return config{
		workspaceID:  &workspaceID,
		taskID:       &taskID,
		teamID:       &teamID,
		riidoAPIHost: &host,
		apiToken:     &token,
		runMutations: &mutate,
		taskFixture:  &fixture,
	}
}

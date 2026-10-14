package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExistingTaskFallbackFindsReadableTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/teams/team-a/components/lists" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []any{map[string]any{"id": "task-a", "componentType": "task"}},
		})
	}))
	defer server.Close()

	got := existingTaskFallback(newAPIClient(server.URL, "token"), "team-a")
	if got.TaskID != "task-a" || got.Scenario.Status != statusPassed {
		t.Fatalf("fallback=%+v", got)
	}
}

func TestExistingTaskFallbackTriesBoardsAfterListsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/teams/team-a/components/lists":
			http.NotFound(w, r)
		case "/teams/team-a/components/boards":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"columns": []any{map[string]any{"cards": []any{
					map[string]any{"id": "task-b", "componentType": "task"},
				}}},
			})
		default:
			t.Fatalf("unexpected request %s", r.URL.String())
		}
	}))
	defer server.Close()

	got := existingTaskFallback(newAPIClient(server.URL, "token"), "team-a")
	if got.TaskID != "task-b" || got.Scenario.Endpoint != "/teams/team-a/components/boards" {
		t.Fatalf("fallback=%+v", got)
	}
}

func TestMarkFixtureFallbackSkipsCreateFailure(t *testing.T) {
	rows := []scenario{{ID: "contract.task.fixture.create", Status: statusFailed, Observed: map[string]any{}}}
	got := markFixtureFallback(rows, taskFallback{TaskID: "task-a", Scenario: scenario{ID: "fallback"}})
	if got[0].Status != statusSkipped || got[0].Observed["fallback_task_id"] != "task-a" {
		t.Fatalf("rows=%+v", got)
	}
}

func TestFirstTaskIDWalksNestedPayload(t *testing.T) {
	payload := map[string]any{"data": []any{map[string]any{"children": []any{
		map[string]any{"id": "project-a", "componentType": "project"},
		map[string]any{"component_id": "task-a", "component_type": "TASK"},
	}}}}
	if got := firstTaskID(payload); got != "task-a" {
		t.Fatalf("task id=%q", got)
	}
}

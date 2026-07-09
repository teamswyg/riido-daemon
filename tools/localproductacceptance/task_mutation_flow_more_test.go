package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskMutationScenariosExerciseFullHappyPath(t *testing.T) {
	var assignmentCalls int
	seen := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen[r.Method+" "+r.URL.Path]++
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/agent-assignments"):
			assignmentCalls++
			writeTaskAction(w, assignmentCalls)
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/thread-stream-subscription"):
			_ = json.NewEncoder(w).Encode(map[string]any{"stream": map[string]any{
				"href": "/events", "event_type": "agent_thread_progress",
			}, "active_thread_filters": []any{map[string]any{"thread_id": "thread_1"}}})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/events"):
			fmt.Fprint(w, "data: asn_1 thread_1\n\ndata: asn_2 thread_2\n\n")
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/threads/thread_1/messages"):
			writeTaskAction(w, 1)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/stop"):
			writeTaskAction(w, 1)
		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/agent-assignments/"):
			writeTaskAction(w, 1)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	plan := taskMutationPlan{
		TaskID:       "task-a",
		TaskIDSource: "configured",
		CommentBody:  "please continue",
		Candidates: []taskAgentCandidate{
			{AgentID: "agent-a", RuntimeKind: "codex"},
			{AgentID: "agent-b", RuntimeKind: "codex"},
		},
	}
	scenarios := taskMutationScenarios(newAPIClient(server.URL, "tok"), "/base", plan)
	if len(scenarios) != 9 {
		t.Fatalf("scenarios=%#v", scenarios)
	}
	for _, sc := range scenarios {
		if sc.Status != statusPassed {
			t.Fatalf("scenario failed: %#v", sc)
		}
	}
	if scenarios[2].ID != "contract.task.multi_assignment" ||
		scenarios[2].Observed["thread_ids_distinct"] != true {
		t.Fatalf("distinct evidence missing: %#v", scenarios[2])
	}
	if seen["POST /base/tasks/task-a/threads/thread_1/messages"] != 1 {
		t.Fatalf("message endpoint not called: %#v", seen)
	}
}

func writeTaskAction(w http.ResponseWriter, n int) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"agent_id": "agent-a", "assignment_id": fmt.Sprintf("asn_%d", n),
		"thread_id": fmt.Sprintf("thread_%d", n), "run_id": fmt.Sprintf("run_%d", n),
		"work_status": "running", "assignment_state": "leased",
	})
}

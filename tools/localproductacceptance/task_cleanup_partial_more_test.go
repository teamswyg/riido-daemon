package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCleanupPartialTaskAssignmentsStopsAndDeletesFirstAgent(t *testing.T) {
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		fmt.Fprint(w, `{"agent_id":"agent-a","assignment_id":"asn-a",`+
			`"thread_id":"thread-a","run_id":"run-a",`+
			`"work_status":"stopped","assignment_state":"stopped"}`)
	}))
	defer server.Close()

	run := taskAssignmentRun{
		First: scenario{ID: "contract.task.assignment.create.first"},
		Pair:  taskAgentPair{First: taskAgentCandidate{AgentID: "agent-a"}},
	}
	scenarios := cleanupPartialTaskAssignments(newAPIClient(server.URL, "tok"), "/base", "task-a", run)
	if len(scenarios) != 2 {
		t.Fatalf("scenarios=%#v", scenarios)
	}
	for _, sc := range scenarios {
		if sc.Status != statusPassed {
			t.Fatalf("scenario failed: %#v", sc)
		}
	}
	joined := strings.Join(seen, "\n")
	for _, want := range []string{
		"POST /base/tasks/task-a/agent-assignments/agent-a/stop",
		"DELETE /base/tasks/task-a/agent-assignments/agent-a",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %s in:\n%s", want, joined)
		}
	}
}

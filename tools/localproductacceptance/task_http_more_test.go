package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateAssignmentRunContinuesUntilTwoSuccesses(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{"error":"runtime binding missing"}`)
			return
		}
		fmt.Fprintf(w, `{"agent_id":"agent_%d","assignment_id":"asn_%d",`+
			`"thread_id":"thread_%d","run_id":"run_%d",`+
			`"work_status":"running","assignment_state":"leased"}`, calls, calls, calls, calls)
	}))
	defer server.Close()
	plan := taskMutationPlan{
		TaskID: "task_1",
		Candidates: []taskAgentCandidate{
			{AgentID: "agent_a", RuntimeKind: "codex", RuntimeID: "rt_a"},
			{AgentID: "agent_b", RuntimeKind: "codex", RuntimeID: "rt_b"},
			{AgentID: "agent_c", RuntimeKind: "codex", RuntimeID: "rt_c"},
		},
	}
	run := createAssignmentRun(newAPIClient(server.URL, "tok"), "/ai-agent", plan)
	if !run.OK || calls != 3 || len(run.Scenarios) != 3 {
		t.Fatalf("unexpected assignment run: calls=%d run=%#v", calls, run)
	}
	if run.First.ID != "contract.task.assignment.create.first" {
		t.Fatalf("first success was not relabeled: %#v", run.First)
	}
	if run.Second.Observed["candidate_runtime_id"] != "rt_c" {
		t.Fatalf("candidate metadata was not preserved: %#v", run.Second)
	}
}

func TestCleanupAndThreadMessageUseTaskScopedEndpoints(t *testing.T) {
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.EscapedPath())
		fmt.Fprint(w, `{"agent_id":"agent_a","assignment_id":"asn_1",`+
			`"thread_id":"thread_1","run_id":"run_1","work_status":"stopped",`+
			`"assignment_state":"stopped"}`)
	}))
	defer server.Close()
	client := newAPIClient(server.URL, "tok")
	plan := taskMutationPlan{TaskID: "task 1"}
	plan.Pair.First.AgentID = "agent_a"
	plan.Pair.Second.AgentID = "agent_b"
	scenarios := cleanupTaskAssignments(client, "/base", plan)
	if len(scenarios) != 3 || scenarios[0].Status != statusPassed {
		t.Fatalf("unexpected cleanup scenarios: %#v", scenarios)
	}
	assigned := scenario{Observed: map[string]any{"thread_id": "thread_1"}}
	msg := threadMessageScenario(client, "/base", plan, assigned)
	if msg.Status != statusPassed {
		t.Fatalf("unexpected message scenario: %#v", msg)
	}
	joined := strings.Join(seen, "\n")
	for _, want := range []string{"/stop", "/agent-assignments/agent_b",
		"/threads/thread_1/messages"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing endpoint %s in:\n%s", want, joined)
		}
	}
}

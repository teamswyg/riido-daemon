package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSSEReplayScenarioRequiresBothAssignmentEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Fatalf("unexpected accept header: %s", r.Header.Get("Accept"))
		}
		fmt.Fprint(w, "id:1\ndata: asn_1\n\nid:2\ndata: thread_2\n\n")
	}))
	defer server.Close()
	first := scenario{Observed: map[string]any{"assignment_id": "asn_1"}}
	second := scenario{Observed: map[string]any{"thread_id": "thread_2"}}
	sc := sseReplayScenario(newAPIClient(server.URL, "tok"), "/base", first, second)
	if sc.Status != statusPassed || sc.Observed["status_code"] != 200 {
		t.Fatalf("unexpected SSE replay scenario: %#v", sc)
	}
}

func TestSSEReplayScenarioTreatsHTTPErrorAsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "server failed")
	}))
	defer server.Close()
	first := scenario{Observed: map[string]any{"assignment_id": "asn_1"}}
	second := scenario{Observed: map[string]any{"assignment_id": "asn_2"}}
	sc := sseReplayScenario(newAPIClient(server.URL, "tok"), "/base", first, second)
	if sc.Status != statusFailed || sc.Repair == nil {
		t.Fatalf("HTTP stream failure should include repair: %#v", sc)
	}
}

func TestBlockedAndPartialAssignmentScenarios(t *testing.T) {
	run := taskAssignmentRun{Scenarios: []scenario{{
		Status: statusFailed, Repair: runtimeBindingRepair(),
	}}}
	blocked := assignmentBlockedScenarios(run)
	if len(blocked) != 4 || blocked[0].Repair.Class != "ai_agent_runtime_binding_required" {
		t.Fatalf("unexpected blocked scenarios: %#v", blocked)
	}
	empty := cleanupPartialTaskAssignments(apiClient{}, "/base", "task", taskAssignmentRun{})
	if empty != nil {
		t.Fatalf("empty partial cleanup should be nil: %#v", empty)
	}
	skipped := taskSkipped(true, "needs task")
	if len(skipped) != 5 || skipped[0].ID != "contract.task.assignable_agents" {
		t.Fatalf("unexpected skipped task scenarios: %#v", skipped)
	}
}

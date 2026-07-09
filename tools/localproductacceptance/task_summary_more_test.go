package main

import "testing"

func TestSummarizeTaskActionSubscriptionAndReplay(t *testing.T) {
	action := summarizeTaskAction(map[string]any{
		"agent_id": "agent_1", "assignment_id": "asn_1",
		"thread_id": "thread_1", "run_id": "run_1",
		"work_status": "running", "assignment_state": "leased",
		"active_stream": map[string]any{"href": "/events"},
	})
	if action["active_stream_set"] != true || action["assignment_id"] != "asn_1" {
		t.Fatalf("unexpected action summary: %#v", action)
	}
	sub := summarizeSubscription(map[string]any{
		"stream": map[string]any{"href": "/events", "event_type": "agent_thread_progress"},
		"active_thread_filters": []any{"a", "b"},
	})
	if sub["stream_href_present"] != true || sub["active_thread_filters_count"] != 2 {
		t.Fatalf("unexpected subscription summary: %#v", sub)
	}
	sc := scenario{Observed: map[string]any{"assignment_id": "asn_1", "thread_id": "thread_1"}}
	if !replayContainsScenario("event asn_1", sc) {
		t.Fatal("replay should match assignment id")
	}
	if !replayContainsScenario("event thread_1", sc) {
		t.Fatal("replay should match thread id")
	}
}

func TestDistinctAssignmentScenarioGuardsDedupeKeys(t *testing.T) {
	plan := taskMutationPlan{TaskIDSource: "configured"}
	plan.Pair.First.RuntimeKind = "claude"
	plan.Pair.Second.RuntimeKind = "claude"
	first := scenario{Status: statusPassed, Observed: map[string]any{
		"assignment_id": "asn_1", "thread_id": "thread_1", "run_id": "run_1",
	}}
	second := scenario{Status: statusPassed, Observed: map[string]any{
		"assignment_id": "asn_2", "thread_id": "thread_2", "run_id": "run_2",
	}}
	sc := distinctAssignmentScenario(plan, first, second)
	if sc.Status != statusPassed || sc.Observed["frontend_dedupe_key"] != "thread_id" {
		t.Fatalf("expected passed multi assignment contract, got %#v", sc)
	}
	second.Observed["run_id"] = "run_1"
	sc = distinctAssignmentScenario(plan, first, second)
	if sc.Status != statusFailed || sc.FailureSummary == "" {
		t.Fatalf("expected collapsed run failure, got %#v", sc)
	}
	first.Status = statusFailed
	sc = distinctAssignmentScenario(plan, first, second)
	if sc.Repair == nil || sc.Repair.Class == "" {
		t.Fatalf("failed create should include repair: %#v", sc)
	}
}

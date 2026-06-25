package main

import "testing"

func TestQASystemExecutionInventoryCounts(t *testing.T) {
	got := qaSystemScenario()
	counts, ok := got.Observed["execution_counts"].(map[string]any)
	if !ok {
		t.Fatalf("execution counts missing: %+v", got.Observed["execution_counts"])
	}
	if counts["system_automated_count"] != 11 || counts["inference_required_count"] != 0 || counts["total"] != 11 {
		t.Fatalf("unexpected execution counts: %+v", counts)
	}
	ids, ok := counts["inference_required_ids"].([]string)
	if !ok || len(ids) != 0 {
		t.Fatalf("inference ids missing: %+v", counts["inference_required_ids"])
	}
	inference, ok := got.Observed["inference_removed"].(map[string]any)
	if !ok || inference["all_execution_automated"] != true {
		t.Fatalf("execution automation state missing: %+v", got.Observed["inference_removed"])
	}
	if inference["system_automated_count"] != 11 || inference["inference_required_count"] != 0 {
		t.Fatalf("execution counts not surfaced in inference audit: %+v", inference)
	}
}

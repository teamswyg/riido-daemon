package main

import "testing"

func TestWorkflowStatusCounts(t *testing.T) {
	records := []workflowRecord{
		{Status: "metadata_only"},
		{Status: "covered"},
		{Status: "covered"},
	}
	got := workflowStatusCounts(records)
	want := []statusCount{
		{Status: "covered", Count: 2},
		{Status: "metadata_only", Count: 1},
	}
	if len(got) != len(want) {
		t.Fatalf("status count length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("status count[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

package main

import "testing"

func TestSummarizeFailsWhenAnyScenarioFails(t *testing.T) {
	got := summarize([]scenario{
		{ID: "a", Status: statusPassed},
		{ID: "b", Status: statusFailed},
	})
	if got != statusFailed {
		t.Fatalf("status=%q", got)
	}
}

func TestSummarizeReturnsPartialWhenScenarioSkipped(t *testing.T) {
	got := summarize([]scenario{{ID: "a", Status: statusSkipped}})
	if got != statusPartial {
		t.Fatalf("status=%q", got)
	}
}

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

func TestSummarizeTreatsPartialScenarioAsPartial(t *testing.T) {
	got := summarize([]scenario{{ID: "local.qa.dsl_system_audit", Status: statusPartial}})
	if got != statusPartial {
		t.Fatalf("summarize partial scenario = %q, want %q", got, statusPartial)
	}
}

package agentexecutionevidence

import (
	"slices"
	"testing"
)

func assertRequiredRisks(t *testing.T, seen map[string]bool) {
	t.Helper()
	for _, risk := range requiredRisks {
		if !seen[risk] {
			t.Fatalf("manifest missing required risk evidence %q", risk)
		}
	}
}

func assertRequiredRemainingBoundaries(t *testing.T, seen map[string]bool) {
	t.Helper()
	for _, id := range requiredRemainingBoundaries {
		if !seen[id] {
			t.Fatalf("remaining boundary %q must stay explicit", id)
		}
	}
}

func assertRiskKnown(t *testing.T, risk string) {
	t.Helper()
	if !slices.Contains(requiredRisks, risk) {
		t.Fatalf("unexpected risk evidence %q", risk)
	}
}

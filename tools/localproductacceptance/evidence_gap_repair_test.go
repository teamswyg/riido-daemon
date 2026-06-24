package main

import (
	"path/filepath"
	"testing"
)

func TestEvidenceGapScenarioUsesScenarioRepair(t *testing.T) {
	manualOut := filepath.Join(t.TempDir(), "manual-qa-evidence.json")
	screenshots := filepath.Join(t.TempDir(), "screenshots", "ai-agent-product-acceptance")
	got := evidenceGapScenario([]scenario{{
		ID:     "local.qa.figma_refresh_gate",
		Status: statusPartial,
		Repair: &repair{Class: "figma_refresh_required", Summary: "Refresh Figma evidence."},
	}}, config{manualOut: &manualOut, screenshots: &screenshots})

	closed := got.Observed["closed_loop_candidates"].([]evidenceGapCandidate)
	if !hasEvidenceGapCandidate(closed, "repair-local.qa.figma_refresh_gate") {
		t.Fatalf("repair candidate missing: %+v", closed)
	}
}

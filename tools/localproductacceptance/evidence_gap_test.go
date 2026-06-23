package main

import (
	"path/filepath"
	"testing"
)

func TestEvidenceGapScenarioRecordsManualAndUploadCandidates(t *testing.T) {
	manualOut := filepath.Join(t.TempDir(), "manual-qa-evidence.json")
	screenshots := filepath.Join(t.TempDir(), "screenshots", "ai-agent-product-acceptance")
	got := evidenceGapScenario([]scenario{
		{ID: "contract.api.bootstrap", Status: statusSkipped},
		{ID: "figma.runtime.detail", Status: statusPassed},
	}, config{manualOut: &manualOut, screenshots: &screenshots})

	if got.ID != "local.qa.evidence_gap_candidates" || got.Status != statusPassed {
		t.Fatalf("gap scenario = %+v", got)
	}
	manual := got.Observed["manual_evidence"].(map[string]any)
	if manual["present"].(bool) || !manual["gap"].(bool) {
		t.Fatalf("manual gap not recorded: %+v", manual)
	}
	capture := got.Observed["feature_ui_capture"].(map[string]any)
	if capture["covered_by_upload_dir"].(bool) {
		t.Fatalf("feature capture should be outside upload dir: %+v", capture)
	}
	candidates := got.Observed["candidates"].([]map[string]string)
	if len(candidates) != 5 || candidates[0]["id"] != "manual-evidence-file" {
		t.Fatalf("candidates = %+v", candidates)
	}
}

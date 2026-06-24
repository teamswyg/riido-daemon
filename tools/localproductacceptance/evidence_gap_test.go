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

	if got.ID != "local.qa.evidence_gap_candidates" || got.Status != statusPartial {
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
	if len(candidates) < 4 || candidates[0]["id"] != "manual-evidence-file" {
		t.Fatalf("candidates = %+v", candidates)
	}
	closed := got.Observed["closed_loop_candidates"].([]evidenceGapCandidate)
	if len(closed) != got.Observed["closed_loop_candidate_count"] {
		t.Fatalf("candidate count mismatch: %+v", got.Observed)
	}
	if closed[len(closed)-1].SourceScenario != "contract.api.bootstrap" {
		t.Fatalf("scenario candidate missing: %+v", closed)
	}
}

func TestEvidenceGapScenarioPromotesInferredExecutionLoops(t *testing.T) {
	manualOut := filepath.Join(t.TempDir(), "manual-qa-evidence.json")
	screenshots := filepath.Join(t.TempDir(), "screenshots", "ai-agent-product-acceptance")
	got := evidenceGapScenario([]scenario{{
		ID:     "local.qa.dsl_system_audit",
		Status: statusPartial,
		Observed: map[string]any{"execution_counts": map[string]any{
			"inference_required_ids": []any{"browser-meaning-qa"},
		}},
	}}, config{manualOut: &manualOut, screenshots: &screenshots})

	closed := got.Observed["closed_loop_candidates"].([]evidenceGapCandidate)
	if !hasEvidenceGapCandidate(closed, "close-inferred-browser-meaning-qa") {
		t.Fatalf("inferred candidate missing: %+v", closed)
	}
}

func hasEvidenceGapCandidate(candidates []evidenceGapCandidate, id string) bool {
	for _, candidate := range candidates {
		if candidate.ID == id {
			return true
		}
	}
	return false
}

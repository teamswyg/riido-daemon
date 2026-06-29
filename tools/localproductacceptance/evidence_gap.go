package main

func evidenceGapScenario(items []scenario, cfg config) scenario {
	summary := summarizeEvidenceGaps(items)
	manualPresent := localFileExists(*cfg.manualOut)
	capturePresent := localFileExists(featureUICapturePath)
	captureUploadCovered := captureCoveredByUploadDir(*cfg.screenshots)
	candidates := evidenceGapCandidates(items, manualPresent, capturePresent, captureUploadCovered)
	status := statusPassed
	if len(candidates) > 0 {
		status = statusPartial
	}
	return scenario{
		ID:     "local.qa.evidence_gap_candidates",
		Status: status,
		Observed: map[string]any{
			"manual_evidence": map[string]any{
				"path":    *cfg.manualOut,
				"present": manualPresent,
				"gap":     !manualPresent,
			},
			"feature_ui_capture": map[string]any{
				"path":                     featureUICapturePath,
				"present":                  capturePresent,
				"covered_by_upload_dir":    captureUploadCovered,
				"current_upload_dir":       *cfg.screenshots,
				"candidate_s3_artifact":    "screenshots/contract-lab/feature-ui-manual-pass.png",
				"needs_upload_loop_update": capturePresent && !captureUploadCovered,
			},
			"skipped_contract_scenarios":      summary.Skipped,
			"figma_without_visual_screenshot": summary.FigmaWithoutScreenshot,
			"closed_loop_candidate_count":     len(candidates),
			"closed_loop_candidates":          candidates,
			"candidates":                      legacyCandidateRows(candidates),
		},
	}
}

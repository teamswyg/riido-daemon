package main

import (
	"os"
	"strings"
)

const featureUICapturePath = ".riido-local/screenshots/contract-lab/feature-ui-manual-pass.png"

func evidenceGapScenario(items []scenario, cfg config) scenario {
	skipped := make([]string, 0)
	withoutScreenshots := make([]string, 0)
	for _, item := range items {
		if item.Status == statusSkipped {
			skipped = append(skipped, item.ID)
		}
		if strings.HasPrefix(item.ID, "figma.") && item.Screenshot == "" {
			withoutScreenshots = append(withoutScreenshots, item.ID)
		}
	}
	manualPresent := localFileExists(*cfg.manualOut)
	capturePresent := localFileExists(featureUICapturePath)
	captureUploadCovered := strings.HasPrefix(featureUICapturePath, strings.TrimRight(*cfg.screenshots, "/")+"/")
	return scenario{
		ID:     "local.qa.evidence_gap_candidates",
		Status: statusPassed,
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
			"skipped_contract_scenarios":      skipped,
			"figma_without_visual_screenshot": withoutScreenshots,
			"candidates": []map[string]string{
				{"id": "manual-evidence-file", "reason": "Manual QA state is browser-local until exported.", "next_evidence": "Create .riido-local/evidence/manual-qa-evidence.json before S3 upload."},
				{"id": "contract-lab-capture-upload", "reason": "Feature UI capture exists outside the current product screenshot upload dir.", "next_evidence": "Upload screenshots/contract-lab or move generated captures under the uploaded screenshot root."},
				{"id": "browser-interaction-runner", "reason": "DSL declares interactions, but localproductacceptance does not replay them by itself.", "next_evidence": "Add a small browser QA runner that emits interaction JSON and PNG."},
				{"id": "runtime-detail-golden", "reason": "Runtime detail has Figma intent evidence but no visual golden screenshot.", "next_evidence": "Capture node 1179:27360 as a golden reference."},
				{"id": "contract-probe-inputs", "reason": "Skipped API contracts need token/workspace/task/agent ids.", "next_evidence": "Run with storage-state plus workspace/task mutation inputs."},
			},
		},
	}
}

func localFileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

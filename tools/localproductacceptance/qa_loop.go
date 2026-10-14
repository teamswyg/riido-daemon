package main

import "time"

func qaLoopScenario(validFor time.Duration, figmaManifest, labOut, manualOut string) scenario {
	return scenario{
		ID:     "local.qa.loop.freshness",
		Status: statusPassed,
		Observed: map[string]any{
			"valid_for_seconds":   int(validFor.Seconds()),
			"figma_manifest":      figmaManifest,
			"react_lab_html":      labOut,
			"manual_evidence":     manualOut,
			"manual_s3_artifact":  "manual-qa-evidence.json",
			"manual_upload_owner": "localqarunner",
			"loop":                "Figma intent entries, daemon-contract evidence, and human manual QA evidence are regenerated or exported into the React lab; dashboard rows inherit the evidence expires_at freshness window.",
			"refresh_owner":       "qa-codex",
		},
	}
}

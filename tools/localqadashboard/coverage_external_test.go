package main

import "testing"

func TestBuildCoverageUsesExternalEvidence(t *testing.T) {
	rows, summary := buildCoverage(
		coverageManifest{Scenarios: []coverageScenario{{
			ID: "product.login", Evidence: "external",
		}}},
		providerEvidenceFile{},
		externalEvidenceFile{Scenarios: []externalScenario{{
			ID: "product.login", Status: "skipped",
			Repair: &repairEvidence{Class: "credential_required"},
		}}},
	)

	if summary.Skipped != 1 || rows[0].Repair.Class != "credential_required" {
		t.Fatalf("external coverage not applied: %+v %+v", summary, rows[0])
	}
}

package main

import (
	"testing"
	"time"
)

func TestQALoopScenarioRecordsFreshnessInputs(t *testing.T) {
	scenario := qaLoopScenario(24*time.Hour, "docs/figma/entries.riido.json", ".riido-local/contract-lab/index.html", ".riido-local/evidence/manual-qa-evidence.json")
	if scenario.ID != "local.qa.loop.freshness" || scenario.Status != statusPassed {
		t.Fatalf("scenario=%+v", scenario)
	}
	if scenario.Observed["valid_for_seconds"] != 86400 {
		t.Fatalf("valid_for_seconds=%v", scenario.Observed["valid_for_seconds"])
	}
	if scenario.Observed["figma_manifest"] == "" || scenario.Observed["react_lab_html"] == "" {
		t.Fatalf("observed paths missing: %+v", scenario.Observed)
	}
	if scenario.Observed["manual_evidence"] == "" || scenario.Observed["manual_s3_artifact"] != "manual-qa-evidence.json" {
		t.Fatalf("manual evidence metadata missing: %+v", scenario.Observed)
	}
}

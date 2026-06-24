package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyProductEvidenceRollsUpClosedLoopCandidates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "product.json")
	body := `{"scenarios":[{"observed":{"closed_loop_candidates":[` +
		`{"id":"close-x","class":"failed_probe","reason":"x","next_evidence":"y",` +
		`"evidence_graph":{"observation":"x","hypothesis":"h","change":"c",` +
		`"verifier":"v","evidence":"e","decision":"d","next_loop":"n"}}]}}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config{productEvidence: &path}
	evidence := runEvidence{CoverageStatus: statusPassed}
	if err := applyProductEvidence(".", cfg, &evidence); err != nil {
		t.Fatal(err)
	}
	if len(evidence.ClosedLoops) != 1 || evidence.CoverageStatus != statusPartial {
		t.Fatalf("closed loops=%+v coverage=%s", evidence.ClosedLoops, evidence.CoverageStatus)
	}
	if evidence.ClosedLoops[0].Graph.NextLoop == "" {
		t.Fatalf("candidate graph missing: %+v", evidence.ClosedLoops[0])
	}
}

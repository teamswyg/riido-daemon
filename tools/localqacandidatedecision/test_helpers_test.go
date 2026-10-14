package main

import (
	"os"
	"path/filepath"
	"testing"
)

func testManifest() manifest {
	return manifest{
		SchemaVersion:    manifestSchema,
		ID:               requiredID,
		Title:            "Local QA Candidate Decision",
		GeneratedDoc:     "docs/decision.md",
		Workflow:         ".github/workflows/local-qa-runner.yml",
		EvidenceArtifact: "local-qa-candidate-decision",
		EvidenceTool:     "tools/localqacandidatedecision",
		Assertions:       []string{"candidate decisions are complete"},
		Loop: evidenceLoop{
			Observation: "o", Hypothesis: "h", Execute: "e",
			Evaluate: "v", Retrospective: "r",
		},
		Decisions: []decisionRecord{testDecision("close-x")},
	}
}

func testDecision(id string) decisionRecord {
	return decisionRecord{
		CandidateID: id, Disposition: "triage_required", Priority: "P1",
		Owner: "qa-loop", NextLoop: "local-qa-gap-to-candidate",
		NextArtifact: "claim_binding", ReviewBy: "2026-12-31", Reason: "needs adoption",
	}
}

func writeCandidate(t *testing.T, dir, body string) string {
	t.Helper()
	path := filepath.Join(dir, "candidate.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func testEvidenceGraphJSON() string {
	return `"evidence_graph":{"observation":"observed gap","hypothesis":"closed loop can prevent recurrence","change":"bind claim","verifier":"candidate-decision","evidence":"candidate.json","decision":"triage","next_loop":"local-qa-gap-to-candidate"}`
}

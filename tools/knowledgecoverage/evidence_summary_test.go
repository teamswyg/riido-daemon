package main

import "testing"

func TestBuildEvidenceExposesTopLevelSummary(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, ".github/workflows/docs.yml", "run: go run ./tools/example -check-doc\n")
	docs := []docClass{{
		Path:      "docs/generated.md",
		Kind:      "generated",
		Generator: "go run ./tools/example -write-doc",
	}}
	got := buildEvidence(root, fixtureManifest(), docs, []string{"fixture problem"})
	if got.ProblemCount != 1 || got.Status != "failed" {
		t.Fatalf("problem summary = %+v", got)
	}
	if got.GeneratedOriginCount != 1 ||
		got.GeneratedWorkflowCovered != 1 ||
		got.GeneratedWorkflowMissing != 0 {
		t.Fatalf("generated workflow summary = %+v", got)
	}
}

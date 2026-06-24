package main

import "testing"

func TestChangedFileEvidenceIncludesProblemDetails(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "go.mod", "module fixture\n")
	writeFixture(t, root, "changed.txt", "code.go\n")
	writeFixture(t, root, defaultManifest, fixtureManifest())

	summary := changedCheck(root, mustRegistry(t, root), "changed.txt")
	if len(summary.ProblemDetails) != 1 {
		t.Fatalf("problem details = %d, want 1", len(summary.ProblemDetails))
	}
	detail := summary.ProblemDetails[0]
	if detail.ClaimID != "claim_binding" || len(detail.RequiredEvidence) == 0 {
		t.Fatalf("bad problem detail: %+v", detail)
	}
}

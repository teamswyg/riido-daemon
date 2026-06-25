package main

import (
	"strings"
	"testing"
)

func TestChangedFileEvidenceIncludesProblemDetails(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "go.mod", "module fixture\n")
	writeFixture(t, root, "changed.txt", "code.go\n")
	writeFixture(t, root, defaultManifest, fixtureManifest())

	summary := changedCheck(root, mustRegistry(t, root), "changed.txt", "")
	if len(summary.ProblemDetails) != 1 {
		t.Fatalf("problem details = %d, want 1", len(summary.ProblemDetails))
	}
	detail := summary.ProblemDetails[0]
	if detail.ClaimID != "claim_binding" || len(detail.RequiredEvidence) == 0 {
		t.Fatalf("bad problem detail: %+v", detail)
	}
}

func TestChangedClaimTextRequiresBoundEvidenceChange(t *testing.T) {
	root := t.TempDir()
	writeLoopRegistryFixture(t, root)
	previous := fixtureManifest()
	current := strings.Replace(previous, "Claims bind code", "Claims strictly bind code", 1)
	writeFixture(t, root, "previous.json", previous)
	writeFixture(t, root, defaultManifest, current)
	writeFixture(t, root, "changed.txt", defaultManifest+"\n")

	summary := changedCheck(root, mustRegistry(t, root), "changed.txt", "previous.json")
	if len(summary.ProblemDetails) != 1 {
		t.Fatalf("problem details = %d, want 1", len(summary.ProblemDetails))
	}
	if summary.ProblemDetails[0].Reason != "business claim changed without bound code/doc/test evidence" {
		t.Fatalf("bad reason: %+v", summary.ProblemDetails[0])
	}
}

func TestChangedClaimTextPassesWithVerifierEvidenceChange(t *testing.T) {
	root := t.TempDir()
	writeLoopRegistryFixture(t, root)
	previous := fixtureManifest()
	current := strings.Replace(previous, "Claims bind code", "Claims strictly bind code", 1)
	writeFixture(t, root, "previous.json", previous)
	writeFixture(t, root, defaultManifest, current)
	writeFixture(t, root, "changed.txt", defaultManifest+"\ncode_test.go\n")

	summary := changedCheck(root, mustRegistry(t, root), "changed.txt", "previous.json")
	if len(summary.ProblemDetails) != 0 {
		t.Fatalf("problem details = %+v, want none", summary.ProblemDetails)
	}
}

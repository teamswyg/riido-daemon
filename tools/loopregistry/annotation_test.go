package main

import (
	"strings"
	"testing"
)

func TestGitHubAnnotationExplainsClaimEvidence(t *testing.T) {
	problem := changedProblem{
		ClaimID:          "claim_binding",
		Reason:           "runtime files changed without bound doc/verifier evidence",
		ChangedFiles:     []string{"code.go"},
		RequiredEvidence: []string{"doc.md", "code_test.go"},
	}
	got := githubAnnotation(problem)
	for _, want := range []string{
		"::error file=code.go",
		"title=Loop Registry Claim Binding",
		"claim claim_binding",
		"Update one bound evidence file: doc.md, code_test.go",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("annotation missing %q in %s", want, got)
		}
	}
}

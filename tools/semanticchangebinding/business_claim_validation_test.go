package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBusinessClaimRequiresExecutablePeers(t *testing.T) {
	binding := Binding{ID: "claim-x", Claim: "claim", ClaimClass: businessClaimClass}
	problems := validateBusinessClaim(binding)
	for _, want := range []string{"code path", "test path", "documentation path", "generated docs", "evidence ids"} {
		if !containsProblem(problems, want) {
			t.Fatalf("expected business claim problem containing %q, got %#v", want, problems)
		}
	}
}

func TestBusinessClaimEvidenceSummary(t *testing.T) {
	repo := fixtureRepo(t)
	out := filepath.Join(t.TempDir(), "semantic.json")
	err := run(context.Background(), options{
		Repo: repo, Manifest: manifestPath(), ChangedFiles: businessClaimPeerFiles(), EvidenceOut: out,
	})
	if err != nil {
		t.Fatalf("expected business claim peers to pass, got %v", err)
	}
	var evidence Evidence
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.BusinessClaims.Count != 1 || evidence.BusinessClaims.VerifiedCount != 1 {
		t.Fatalf("unexpected business claim summary: %#v", evidence.BusinessClaims)
	}
}

func containsProblem(problems []problem, value string) bool {
	for _, problem := range problems {
		if strings.Contains(problem.Message, value) {
			return true
		}
	}
	return false
}

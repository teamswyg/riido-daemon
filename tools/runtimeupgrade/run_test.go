package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesVerifiedEvidenceWithReservedRows(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	out := filepath.Join(repo, "out.json")
	err := run(t.Context(), options{
		Repo: repo, Manifest: defaultManifest,
		WriteDoc: true, CheckDoc: true, EvidenceOut: out,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunDetectsReservedRuleWithSourceChecks(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	body := fixtureManifestSource
	body = strings.Replace(body, `"required_evidence": "test"`, `"source_checks": ["source"], "required_evidence": "test"`, 1)
	writeFile(t, repo, defaultManifest, body)
	if err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest}); err == nil {
		t.Fatal("expected reserved rule with source checks to fail")
	}
}

func TestRunDetectsForbiddenClaim(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	writeFile(t, repo, "docs/claims.md", "RuntimePinViolated\n")
	if err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest}); err == nil {
		t.Fatal("expected forbidden implemented claim")
	}
}

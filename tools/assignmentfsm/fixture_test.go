package main

import "testing"

func TestAssignmentFSMFixtureUsesContractsSource(t *testing.T) {
	manifest, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.SourcePackage != "github.com/teamswyg/riido-contracts/assignment" {
		t.Fatalf("source package = %q", manifest.SourcePackage)
	}
	if len(manifest.ForbiddenDocTokens) == 0 {
		t.Fatal("expected stale-state forbidden tokens")
	}
}

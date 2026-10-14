package main

import (
	"path/filepath"
	"testing"
)

func TestRunWritesVerifiedEvidence(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	out := filepath.Join(repo, "out.json")
	err := run(t.Context(), options{
		Repo: repo, Manifest: defaultManifest, WriteDoc: true,
		WriteScript: true, CheckDoc: true, CheckScript: true, EvidenceOut: out,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunDetectsScriptBehaviorDrift(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	writeFile(t, repo, "scripts/verify-branch.sh", "#!/usr/bin/env bash\nexit 0\n", 0o755)
	err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest})
	if err == nil {
		t.Fatal("expected rejected example to fail")
	}
}

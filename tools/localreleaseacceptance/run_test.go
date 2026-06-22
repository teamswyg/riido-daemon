package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunWritesFreshInstallEvidence(t *testing.T) {
	repo := fixtureRepo(t)
	out := filepath.Join(t.TempDir(), "release.json")
	err := run(t.Context(), options{Repo: repo, EvidenceOut: out, ValidFor: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	var evidence evidenceFile
	if err := json.Unmarshal(readFile(t, out), &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.Status != statusPassed || evidence.Scenarios[0].ID != "release.fresh.install" {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
	if evidence.Artifacts.VersionOutput != "riido version v-local-qa" {
		t.Fatalf("version output=%q", evidence.Artifacts.VersionOutput)
	}
}

func fixtureRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	script := filepath.Join(repo, "scripts", "install-riido-daemon.sh")
	if err := os.MkdirAll(filepath.Dir(script), 0o755); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("../../scripts/install-riido-daemon.sh")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script, source, 0o755); err != nil {
		t.Fatal(err)
	}
	return repo
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

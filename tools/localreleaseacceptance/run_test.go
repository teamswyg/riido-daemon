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
	api := releaseAPIServer(t, releaseBody(expectedReleaseAsset(), "SHA256SUMS"))
	out := filepath.Join(t.TempDir(), "release.json")
	opts := options{Repo: repo, EvidenceOut: out, ValidFor: time.Hour, ReleaseAPIURL: api}
	err := run(t.Context(), opts)
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
	if evidence.Scenarios[1].ID != "release.github.latest_assets" {
		t.Fatalf("latest release scenario missing: %+v", evidence.Scenarios)
	}
	if evidence.Artifacts.VersionOutput != "riido version v-local-qa" {
		t.Fatalf("version output=%q", evidence.Artifacts.VersionOutput)
	}
	if evidence.Artifacts.LatestReleaseTag != "v-test" {
		t.Fatalf("latest release tag=%q", evidence.Artifacts.LatestReleaseTag)
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

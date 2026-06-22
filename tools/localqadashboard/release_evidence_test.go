package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReleaseEvidenceScenarios(t *testing.T) {
	path := filepath.Join(t.TempDir(), "release.json")
	body := `{"scenarios":[{"id":"release.fresh.install","status":"passed"}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := releaseEvidenceScenarios(path)
	if len(got) != 1 || got[0].ID != "release.fresh.install" {
		t.Fatalf("scenarios=%+v", got)
	}
}

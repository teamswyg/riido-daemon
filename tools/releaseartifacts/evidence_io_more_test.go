package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseArtifactEvidenceAndProblemHelpers(t *testing.T) {
	m := manifest{
		ID:               "release",
		GeneratedDoc:     "docs/release.md",
		EvidenceArtifact: "evidence.json",
		DetailDocs:       []detailDoc{{Path: "a"}, {Path: "b"}, {Path: "c"}, {Path: "d"}},
		Targets:          []target{{GOOS: "darwin", GOARCH: "arm64", Format: "tar.gz"}},
	}
	problems := []problem{{Message: "missing workflow"}}
	checks := []checkResult{{Name: "script", File: "scripts/release.sh", Pass: false}}
	ev := buildEvidence(m, problems, checks)
	if ev.Status != "failed" || ev.GeneratedDocs[4] != "d" || ev.Targets[0] != "darwin/arm64/tar.gz" {
		t.Fatalf("unexpected evidence=%+v", ev)
	}
	if !strings.Contains(problemError(problems).Error(), "missing workflow") {
		t.Fatal("problem error should include messages")
	}
	if got := problemMessages(problems); len(got) != 1 || got[0] != "missing workflow" {
		t.Fatalf("problem messages=%+v", got)
	}
}

func TestReleaseArtifactsIOAndStrictManifestParsing(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "docs", "release.md")
	if err := writeText(textPath, "doc"); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(textPath); err != nil || string(data) != "doc" {
		t.Fatalf("text=%q err=%v", data, err)
	}
	jsonPath := filepath.Join(dir, "evidence.json")
	if err := writeJSON(jsonPath, map[string]string{"id": "release"}); err != nil {
		t.Fatal(err)
	}
	unknown := filepath.Join(dir, "unknown.json")
	if err := os.WriteFile(unknown, []byte(`{"unknown":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadManifest("", unknown); err == nil {
		t.Fatal("unknown manifest fields must fail")
	}
	trailing := filepath.Join(dir, "trailing.json")
	if err := os.WriteFile(trailing, []byte(`{} {}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadManifest("", trailing); err == nil {
		t.Fatal("trailing JSON values must fail")
	}
}

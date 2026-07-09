package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestAndWriteJSONHelpers(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadManifest(bad); err == nil || !strings.Contains(err.Error(), "parse manifest") {
		t.Fatalf("expected parse error, got %v", err)
	}
	out := filepath.Join(dir, "evidence.json")
	if err := writeJSON(out, map[string]string{"status": "ok"}); err != nil {
		t.Fatal(err)
	}
	raw, readErr := os.ReadFile(out)
	if readErr != nil || !strings.HasSuffix(string(raw), "\n") {
		t.Fatalf("json=%q err=%v", raw, readErr)
	}
}

func TestMaybeWriteOrCheckDocBranches(t *testing.T) {
	dir := t.TempDir()
	m := manifest{
		Title:            "Provider QA",
		GeneratedDoc:     "doc.md",
		EvidenceArtifact: "evidence.json",
		Providers:        []provider{{ID: "codex", DisplayName: "Codex"}},
	}
	if err := maybeWriteOrCheckDoc(dir, m, false, false); err != nil {
		t.Fatalf("disabled check returned %v", err)
	}
	if err := maybeWriteOrCheckDoc(dir, m, true, false); err != nil {
		t.Fatal(err)
	}
	if err := maybeWriteOrCheckDoc(dir, m, false, true); err != nil {
		t.Fatalf("fresh doc returned %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "doc.md"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := maybeWriteOrCheckDoc(dir, m, false, true)
	if err == nil || !strings.Contains(err.Error(), "generated doc drift") {
		t.Fatalf("expected doc drift, got %v", err)
	}
}

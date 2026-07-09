package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathProblemAndIOHelpers(t *testing.T) {
	dir := t.TempDir()
	if got := resolvePath(dir, "/tmp/file"); got != "/tmp/file" {
		t.Fatalf("absolute path resolved to %q", got)
	}
	if got := joinProblems([]string{"one", "two"}); got != "- one\n- two\n" {
		t.Fatalf("unexpected problems %q", got)
	}
	bad := filepath.Join(dir, "bad.json")
	mustWrite(t, bad, "{")
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
		Title:            "Runtime Secret",
		GeneratedDoc:     "doc.md",
		EvidenceArtifact: "evidence.json",
		PrivateOwner:     "private-infra",
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
	mustWrite(t, filepath.Join(dir, "doc.md"), "stale")
	err := maybeWriteOrCheckDoc(dir, m, false, true)
	if err == nil || !strings.Contains(err.Error(), "generated doc drift") {
		t.Fatalf("expected doc drift, got %v", err)
	}
}

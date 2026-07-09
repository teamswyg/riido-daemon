package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReportsRedactionDrift(t *testing.T) {
	dir := t.TempDir()
	writeRedactionDriftFile(t, dir, "docs/20-domain/security/example.md", "redaction required")
	err := run(dir)
	if err == nil {
		t.Fatal("expected drift error")
	}
	for _, want := range []string{"redaction drift:", "- docs/20-domain/security/example.md"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err.Error(), want)
		}
	}
}

func TestRunWrapsScanErrors(t *testing.T) {
	err := run(filepath.Join(t.TempDir(), "missing"))
	if err == nil || !strings.Contains(err.Error(), "scan redaction docs") {
		t.Fatalf("expected scan error, got %v", err)
	}
}

func TestJoinProblemsFormatsBullets(t *testing.T) {
	got := joinProblems([]string{"one", "two"})
	if got != "- one\n- two\n" {
		t.Fatalf("joined problems = %q", got)
	}
}

func TestRedactionSSOTRequiresCanonicalMarkers(t *testing.T) {
	problems := validateRedactionSSOT("docs/20-domain/security-redaction/markers.md", "empty")
	if len(problems) != 2 {
		t.Fatalf("problems = %#v, want two missing marker problems", problems)
	}
	if got := validateRedactionSSOT("docs/20-domain/security-redaction/other.md", "empty"); len(got) != 0 {
		t.Fatalf("non-marker page problems = %#v", got)
	}
}

func writeRedactionDriftFile(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

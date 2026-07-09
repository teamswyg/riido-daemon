package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteJSONReportsMarshalAndParentErrors(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "not-dir"), []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := writeJSON(filepath.Join(root, "not-dir", "out.json"), map[string]string{"ok": "true"})
	if err == nil {
		t.Fatal("expected parent creation error")
	}

	err = writeJSON(filepath.Join(root, "out.json"), map[string]any{"bad": func() {}})
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("expected marshal error, got %v", err)
	}
}

func TestTimestampSlugFallsBackForInvalidTime(t *testing.T) {
	if got := timestampSlug("2026-06-22 07:32:49"); got != "20260622 073249" {
		t.Fatalf("fallback slug=%q", got)
	}
}

func TestTailTrimsAndLimitsOutput(t *testing.T) {
	if got := tail("\nabc\n", 10); got != "abc" {
		t.Fatalf("trimmed tail=%q", got)
	}
	if got := tail("abcdef", 3); got != "def" {
		t.Fatalf("limited tail=%q", got)
	}
}

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateCandidateMetadataCoversClosedCandidateAndBadDates(t *testing.T) {
	closed := loopEntry{ID: "closed", Kind: "closed-loop"}
	if problems := validateCandidateMetadata(closed); len(problems) != 0 {
		t.Fatalf("closed loop should not require candidate metadata: %v", problems)
	}

	missing := loopEntry{ID: "candidate", Kind: "candidate"}
	if got := strings.Join(validateCandidateMetadata(missing), "\n"); !strings.Contains(got, "candidate_created_at") {
		t.Fatalf("expected missing candidate metadata problem, got %q", got)
	}

	badDate := loopEntry{
		ID:                 "candidate",
		Kind:               "candidate",
		CandidateCreatedAt: "2026/07/09",
		PromotionTarget:    "closed-loop",
	}
	if got := strings.Join(validateCandidateMetadata(badDate), "\n"); !strings.Contains(got, "must be YYYY-MM-DD") {
		t.Fatalf("expected bad date problem, got %q", got)
	}

	valid := badDate
	valid.CandidateCreatedAt = "2026-07-09"
	if problems := validateCandidateMetadata(valid); len(problems) != 0 {
		t.Fatalf("valid candidate metadata should pass: %v", problems)
	}
}

func TestLoopRegistryWritersReportParentAndMarshalErrors(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "not-dir"), []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	textPath := filepath.Join(root, "not-dir", "doc.md")
	if err := writeText(textPath, "doc"); err == nil {
		t.Fatal("expected text parent creation error")
	}

	err := writeJSON(filepath.Join(root, "bad.json"), map[string]any{"bad": func() {}})
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("expected JSON marshal error, got %v", err)
	}
}

func TestRepoPathKeepsAbsoluteAndRootsRelativePaths(t *testing.T) {
	abs := filepath.Join(t.TempDir(), "loop.json")
	if got := repoPath("/repo", abs); got != abs {
		t.Fatalf("absolute path changed: %q", got)
	}
	if got := repoPath("/repo", "docs/loop.json"); got != filepath.Join("/repo", "docs", "loop.json") {
		t.Fatalf("relative path not rooted: %q", got)
	}
	if got := slash(" docs\\loop.json "); got != "docs\\loop.json" {
		t.Fatalf("slash should trim only host separator conversion changed value: %q", got)
	}
}

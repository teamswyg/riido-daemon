package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckDigestOrderingAndCompareEdges(t *testing.T) {
	checks := []sourceCheck{
		{Name: "z", File: "b.go", Contains: []string{"b", "a"}},
		{Name: "a", File: "z.go", Contains: []string{"d"}},
		{Name: "a", File: "a.go", Contains: []string{"c"}},
	}
	got := checkDigests(checks)
	if got[0].Name != "a" || got[0].File != "a.go" {
		t.Fatalf("unexpected first digest: %+v", got[0])
	}
	if strings.Join(got[2].Contains, ",") != "a,b" {
		t.Fatalf("contains not sorted: %+v", got[2].Contains)
	}
	if compareCheckDigest(got[0], got[0]) != 0 {
		t.Fatal("identical digests should compare equal")
	}
	if compareString("b", "a") <= 0 || compareString("a", "b") >= 0 {
		t.Fatal("compareString ordering is inverted")
	}
}

func TestEmitGitHubAnnotationsWritesAllProblems(t *testing.T) {
	restore, stderr := captureStderr(t)
	emitGitHubAnnotations(changedSummary{ProblemDetails: []changedProblem{
		{ClaimID: "c1", Reason: "r1", ChangedFiles: []string{"a.go"}, RequiredEvidence: []string{"a_test.go"}},
		{ClaimID: "c2", Reason: "r2", ChangedFiles: []string{"b.go"}, RequiredEvidence: []string{"b_test.go"}},
	}})
	restore()
	got := stderr()
	if strings.Count(got, "::error file=") != 2 {
		t.Fatalf("expected two annotations, got %q", got)
	}
	if !strings.Contains(got, "claim c1") || !strings.Contains(got, "claim c2") {
		t.Fatalf("missing claim ids in %q", got)
	}
}

func TestWriteJSONAndTextCreateParents(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "nested", "evidence.json")
	if err := writeJSON(jsonPath, map[string]string{"ok": "yes"}); err != nil {
		t.Fatal(err)
	}
	var decoded map[string]string
	if err := json.Unmarshal(mustRead(t, jsonPath), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["ok"] != "yes" {
		t.Fatalf("unexpected json content: %+v", decoded)
	}
	textPath := filepath.Join(dir, "docs", "loop.md")
	if err := writeText(textPath, "loop evidence"); err != nil {
		t.Fatal(err)
	}
	if string(mustRead(t, textPath)) != "loop evidence" {
		t.Fatal("text output mismatch")
	}
}

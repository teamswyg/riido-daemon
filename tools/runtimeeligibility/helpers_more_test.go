package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProblemAndIOHelpers(t *testing.T) {
	if err := problemError(nil); err != nil {
		t.Fatalf("empty problems returned %v", err)
	}
	err := problemError([]problem{{Message: "one"}, {Message: "two"}})
	if err == nil || err.Error() != "one\ntwo" {
		t.Fatalf("unexpected problem error %v", err)
	}
	path := filepath.Join(t.TempDir(), "evidence.json")
	if err := writeJSON(path, map[string]string{"status": "ok"}); err != nil {
		t.Fatal(err)
	}
	raw, readErr := os.ReadFile(path)
	if readErr != nil || !strings.HasSuffix(string(raw), "\n") {
		t.Fatalf("json=%q err=%v", raw, readErr)
	}
}

func TestCheckDocBranches(t *testing.T) {
	dir := t.TempDir()
	if got := checkDoc(options{Repo: dir}, "doc.md", "body"); got != nil {
		t.Fatalf("disabled check returned %#v", got)
	}
	assertRuntimeProblem(
		t,
		checkDoc(options{Repo: dir, CheckDoc: true}, "missing.md", "body"),
		"missing.md",
	)
	path := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(path, []byte("body"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := checkDoc(options{Repo: dir, CheckDoc: true}, "doc.md", "body"); got != nil {
		t.Fatalf("fresh doc returned %#v", got)
	}
	assertRuntimeProblem(
		t,
		checkDoc(options{Repo: dir, CheckDoc: true}, "doc.md", "new"),
		"generated doc drift",
	)
}

func assertRuntimeProblem(t *testing.T, problems []problem, needle string) {
	t.Helper()
	for _, item := range problems {
		if strings.Contains(item.Message, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}

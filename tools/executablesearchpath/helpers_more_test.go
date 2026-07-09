package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProblemAndIOHelpers(t *testing.T) {
	if err := problemError(nil); err != nil {
		t.Fatalf("empty problems returned error %v", err)
	}
	err := problemError([]problem{{Message: "first"}, {Message: "second"}})
	if err == nil || err.Error() != "first\nsecond" {
		t.Fatalf("unexpected joined error %v", err)
	}
	path := filepath.Join(t.TempDir(), "out.json")
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
	missing := checkDoc(options{Repo: dir, CheckDoc: true}, "missing.md", "body")
	assertExecPathProblem(t, missing, "missing.md")
	if err := writeText(filepath.Join(dir, "doc.md"), "body"); err != nil {
		t.Fatal(err)
	}
	if got := checkDoc(options{Repo: dir, CheckDoc: true}, "doc.md", "body"); got != nil {
		t.Fatalf("fresh doc returned %#v", got)
	}
	assertExecPathProblem(
		t,
		checkDoc(options{Repo: dir, CheckDoc: true}, "doc.md", "fresh"),
		"generated doc drift",
	)
}

func assertExecPathProblem(t *testing.T, problems []problem, needle string) {
	t.Helper()
	for _, item := range problems {
		if strings.Contains(item.Message, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}

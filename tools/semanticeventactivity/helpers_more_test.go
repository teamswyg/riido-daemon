package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathProblemAndIOHelpers(t *testing.T) {
	dir := t.TempDir()
	if _, err := cleanRepoPath(dir, "/tmp/file"); err == nil {
		t.Fatal("absolute path accepted")
	}
	if _, err := cleanRepoPath(dir, "../file"); err == nil {
		t.Fatal("escaping path accepted")
	}
	if err := problemError(nil); err != nil {
		t.Fatalf("empty problems returned %v", err)
	}
	err := problemError([]problem{{msg: "first"}, {msg: "second"}})
	if err == nil || err.Error() != "first" {
		t.Fatalf("unexpected problem error %v", err)
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

func TestDocWriteAndCheckBranches(t *testing.T) {
	dir := t.TempDir()
	m := Manifest{Title: "Semantic Events", GeneratedDoc: "doc.md"}
	rendered := render(m)
	if got := checkDoc(options{Repo: dir}, m, rendered); got != nil {
		t.Fatalf("disabled check returned %#v", got)
	}
	assertSemanticProblem(t, checkDoc(options{Repo: dir, CheckDoc: true}, m, rendered), "doc.md")
	if err := maybeWriteDoc(options{Repo: dir, WriteDoc: true}, m, rendered); err != nil {
		t.Fatal(err)
	}
	if got := checkDoc(options{Repo: dir, CheckDoc: true}, m, rendered); got != nil {
		t.Fatalf("fresh doc returned %#v", got)
	}
	if err := os.WriteFile(filepath.Join(dir, "doc.md"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	assertSemanticProblem(t, checkDoc(options{Repo: dir, CheckDoc: true}, m, rendered), "generated doc drift")
}

func assertSemanticProblem(t *testing.T, problems []problem, needle string) {
	t.Helper()
	for _, item := range problems {
		if strings.Contains(item.msg, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}

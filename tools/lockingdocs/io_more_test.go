package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadJSONRejectsTrailingValueAndWriteTextCreatesParents(t *testing.T) {
	dir := t.TempDir()
	trailing := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(trailing, []byte(`{"id":"one"} {"id":"two"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out manifest
	if err := readJSON(trailing, &out); err == nil || !strings.Contains(err.Error(), "trailing") {
		t.Fatalf("expected trailing JSON error, got %v", err)
	}
	path := filepath.Join(dir, "nested", "doc.md")
	if err := writeText(path, "body"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != "body" {
		t.Fatalf("body=%q err=%v", raw, err)
	}
}

func TestMaybeWriteAndCheckDocsBranches(t *testing.T) {
	dir := t.TempDir()
	docs := map[string]string{"docs/out.md": "fresh"}
	if err := maybeWriteDocs(options{Repo: dir}, docs); err != nil {
		t.Fatal(err)
	}
	assertLockingProblem(t, checkDocs(options{Repo: dir, CheckDoc: true}, docs), "docs/out.md")
	if err := maybeWriteDocs(options{Repo: dir, WriteDoc: true}, docs); err != nil {
		t.Fatal(err)
	}
	if got := checkDocs(options{Repo: dir, CheckDoc: true}, docs); got != nil {
		t.Fatalf("fresh docs returned %#v", got)
	}
	if err := writeText(filepath.Join(dir, "docs/out.md"), "stale"); err != nil {
		t.Fatal(err)
	}
	assertLockingProblem(
		t,
		checkDocs(options{Repo: dir, CheckDoc: true}, docs),
		"generated doc drift",
	)
}

func assertLockingProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}

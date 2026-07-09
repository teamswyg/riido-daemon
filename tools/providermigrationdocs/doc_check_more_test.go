package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaybeWriteDocsAndCheckDocsRoundTrip(t *testing.T) {
	root := t.TempDir()
	docs := map[string]string{
		"docs/root.md": "root",
		"docs/page.md": "page",
	}
	if err := maybeWriteDocs(options{Repo: root}, docs); err != nil {
		t.Fatalf("disabled write should not fail: %v", err)
	}
	if _, err := os.Stat(repoPath(root, "docs/root.md")); !os.IsNotExist(err) {
		t.Fatalf("disabled write created a doc: %v", err)
	}
	if err := maybeWriteDocs(options{Repo: root, WriteDoc: true}, docs); err != nil {
		t.Fatalf("write docs: %v", err)
	}
	if problems := checkDocs(options{Repo: root, CheckDoc: true}, docs); len(problems) != 0 {
		t.Fatalf("fresh docs should pass: %v", problems)
	}
	if err := writeText(repoPath(root, "docs/page.md"), "stale"); err != nil {
		t.Fatal(err)
	}
	problems := checkDocs(options{Repo: root, CheckDoc: true}, docs)
	if len(problems) != 1 || !strings.Contains(problems[0], "generated doc drift") {
		t.Fatalf("expected drift problem, got %v", problems)
	}
}

func TestCheckDocsReportsMissingFileAndRepoPath(t *testing.T) {
	root := t.TempDir()
	problems := checkDocs(options{Repo: root, CheckDoc: true}, map[string]string{"missing.md": ""})
	if len(problems) != 1 || !strings.Contains(problems[0], "no such file") {
		t.Fatalf("expected missing file problem, got %v", problems)
	}
	abs := filepath.Join(root, "absolute.md")
	if got := repoPath("ignored", abs); got != abs {
		t.Fatalf("absolute repo path changed: %q", got)
	}
}

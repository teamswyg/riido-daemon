package main

import (
	"os"
	"strings"
	"testing"
)

func TestMaybeWriteDocsAndCheckDocsRoundTrip(t *testing.T) {
	root := t.TempDir()
	docs := map[string]string{
		"docs/root.md":   "root",
		"docs/detail.md": "detail",
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
		t.Fatalf("fresh docs should pass: %+v", problems)
	}
	if err := writeText(repoPath(root, "docs/detail.md"), "stale"); err != nil {
		t.Fatal(err)
	}
	problems := checkDocs(options{Repo: root, CheckDoc: true}, docs)
	if len(problems) != 1 || !strings.Contains(problems[0].Message, "generated doc drift") {
		t.Fatalf("expected drift problem, got %+v", problems)
	}
}

func TestCheckDocsReportsMissingFileAndNormalizeOptions(t *testing.T) {
	root := t.TempDir()
	problems := checkDocs(options{Repo: root, CheckDoc: true}, map[string]string{"missing.md": ""})
	if len(problems) != 1 || !strings.Contains(problems[0].Message, "no such file") {
		t.Fatalf("expected missing file problem, got %+v", problems)
	}
	opts := normalizeOptions(options{})
	if opts.Repo != "." || opts.Manifest != defaultManifest {
		t.Fatalf("default options not applied: %+v", opts)
	}
}

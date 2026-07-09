package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaybeWriteDocsWritesOnlyWhenEnabled(t *testing.T) {
	repo := t.TempDir()
	docs := map[string]string{"docs/a.md": "hello"}
	if err := maybeWriteDocs(options{Repo: repo}, docs); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repo, "docs", "a.md")); !os.IsNotExist(err) {
		t.Fatalf("disabled write should not create doc, err=%v", err)
	}
	if err := maybeWriteDocs(options{Repo: repo, WriteDoc: true}, docs); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(repo, "docs", "a.md")); err != nil ||
		string(got) != "hello" {
		t.Fatalf("written doc = %q err=%v", got, err)
	}
}

func TestCheckDocsReportsMissingAndDriftOnlyWhenEnabled(t *testing.T) {
	repo := t.TempDir()
	docs := map[string]string{"docs/a.md": "expected"}
	if problems := checkDocs(options{Repo: repo}, docs); len(problems) != 0 {
		t.Fatalf("disabled check should not report: %#v", problems)
	}
	problems := checkDocs(options{Repo: repo, CheckDoc: true}, docs)
	if len(problems) != 1 || !strings.Contains(problems[0].Message, "no such file") {
		t.Fatalf("missing doc not reported: %#v", problems)
	}
	if err := maybeWriteDocs(options{Repo: repo, WriteDoc: true}, docs); err != nil {
		t.Fatal(err)
	}
	if problems := checkDocs(options{Repo: repo, CheckDoc: true}, docs); len(problems) != 0 {
		t.Fatalf("current docs should pass: %#v", problems)
	}
	if err := os.WriteFile(filepath.Join(repo, "docs", "a.md"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	problems = checkDocs(options{Repo: repo, CheckDoc: true}, docs)
	if len(problems) != 1 || !strings.Contains(problems[0].Message, "generated doc drift") {
		t.Fatalf("drift not reported: %#v", problems)
	}
}

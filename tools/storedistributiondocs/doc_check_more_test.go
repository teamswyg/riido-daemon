package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaybeWriteDocsAndCheckDocs(t *testing.T) {
	root := t.TempDir()
	docs := map[string]string{"docs/a.md": "A", "docs/b.md": "B"}
	opts := options{Repo: root, WriteDoc: false, CheckDoc: true}
	if err := maybeWriteDocs(opts, docs); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "a.md")); !os.IsNotExist(err) {
		t.Fatalf("write disabled should not create docs, got %v", err)
	}
	opts.WriteDoc = true
	if err := maybeWriteDocs(opts, docs); err != nil {
		t.Fatal(err)
	}
	if problems := checkDocs(opts, docs); len(problems) != 0 {
		t.Fatalf("fresh docs problems = %+v", problems)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "a.md"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	problems := checkDocs(opts, docs)
	if len(problems) != 1 || !strings.Contains(problems[0], "docs/a.md is stale") {
		t.Fatalf("stale problems = %+v", problems)
	}
	if err := os.Remove(filepath.Join(root, "docs", "b.md")); err != nil {
		t.Fatal(err)
	}
	problems = checkDocs(opts, docs)
	if len(problems) != 2 || !strings.Contains(strings.Join(problems, "\n"), "read docs/b.md") {
		t.Fatalf("missing doc problems = %+v", problems)
	}
}

func TestRenderedIfValidSkipsDocsOnProblems(t *testing.T) {
	m := manifest{GeneratedDoc: "docs/root.md"}
	c := contract{Channels: []channel{{ID: "mac-app-store"}}}
	if got := renderedIfValid(m, c, []string{"bad"}); len(got) != 0 {
		t.Fatalf("invalid render should be empty, got %+v", got)
	}
	if got := renderedIfValid(m, c, nil); len(got) == 0 {
		t.Fatal("valid render should produce docs")
	}
}

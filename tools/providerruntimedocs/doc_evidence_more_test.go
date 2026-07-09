package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderedDocsAndEvidence(t *testing.T) {
	m := testManifest()
	docs := renderedDocs(m)
	if len(docs) != 4 {
		t.Fatalf("docs = %d, want root plus three pages", len(docs))
	}
	root := docs[m.GeneratedDoc]
	if !strings.Contains(root, "Compatibility Markers") ||
		!strings.Contains(root, "[one](one.md)") {
		t.Fatalf("root doc missing generated sections:\n%s", root)
	}
	page := docs[m.Pages[0].GeneratedDoc]
	if !strings.Contains(page, "[Back to root]") ||
		!strings.Contains(page, "- detect") {
		t.Fatalf("page doc missing rendered blocks:\n%s", page)
	}
	ev := buildEvidence(m, nil, []string{"drift"})
	if ev.Status != "failed" || ev.PageCount != 4 ||
		len(ev.GeneratedDocs) != 4 || ev.ProblemSummaries[0] != "drift" {
		t.Fatalf("evidence = %+v", ev)
	}
}

func TestWriteAndCheckDocs(t *testing.T) {
	repo := t.TempDir()
	m := testManifest()
	docs := renderedDocs(m)
	opts := options{Repo: repo, WriteDoc: true, CheckDoc: true}
	if err := maybeWriteDocs(opts, docs); err != nil {
		t.Fatalf("write docs: %v", err)
	}
	if problems := checkDocs(opts, docs); len(problems) != 0 {
		t.Fatalf("check written docs: %+v", problems)
	}
	path := repoPath(repo, m.GeneratedDoc)
	if err := os.WriteFile(path, []byte("stale"), 0o644); err != nil {
		t.Fatalf("stale doc: %v", err)
	}
	problems := checkDocs(opts, docs)
	if len(problems) != 1 || !strings.Contains(problems[0], "generated doc drift") {
		t.Fatalf("drift problems: %+v", problems)
	}
}

func TestWriteJSONCreatesIndentedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "evidence.json")
	if err := writeJSON(path, map[string]string{"status": "verified"}); err != nil {
		t.Fatalf("write json: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	if got := string(body); !strings.Contains(got, "\n  \"status\"") ||
		!strings.HasSuffix(got, "\n") {
		t.Fatalf("json body = %q", got)
	}
}

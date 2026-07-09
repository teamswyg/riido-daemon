package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteCheckDocsAndJSON(t *testing.T) {
	repo := t.TempDir()
	docs := map[string]string{"docs/a.md": "fresh"}
	opts := options{Repo: repo, WriteDoc: true, CheckDoc: true}
	if err := maybeWriteDocs(opts, docs); err != nil {
		t.Fatalf("write docs: %v", err)
	}
	if problems := checkDocs(opts, docs); len(problems) != 0 {
		t.Fatalf("check docs: %+v", problems)
	}
	if err := os.WriteFile(repoPath(repo, "docs/a.md"), []byte("stale"), 0o644); err != nil {
		t.Fatalf("stale doc: %v", err)
	}
	if problems := checkDocs(opts, docs); len(problems) != 1 ||
		!strings.Contains(problems[0], "is stale") {
		t.Fatalf("drift problems: %+v", problems)
	}
	out := filepath.Join(repo, "out", "result.json")
	if err := writeJSON(out, map[string]string{"status": "verified"}); err != nil {
		t.Fatalf("write json: %v", err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	if !strings.Contains(string(body), "\n  \"status\"") || !strings.HasSuffix(string(body), "\n") {
		t.Fatalf("json body = %q", body)
	}
}

func TestReadJSONRejectsTrailingJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"x"} {}`), 0o644); err != nil {
		t.Fatalf("write json: %v", err)
	}
	var m manifest
	err := readJSON(path, &m)
	if err == nil || !strings.Contains(err.Error(), "unexpected trailing JSON") {
		t.Fatalf("read json err = %v", err)
	}
}

func TestValidateModelAndRenderedIfValid(t *testing.T) {
	m := testModel()
	if problems := validateModel(m); len(problems) != 0 {
		t.Fatalf("valid model problems = %+v", problems)
	}
	if docs := renderedIfValid(m, nil); len(docs) != 14 {
		t.Fatalf("valid docs = %d", len(docs))
	}
	m.Manifest.ID = ""
	problems := validateModel(m)
	if !hasAgentDesignProblem(problems, "required") {
		t.Fatalf("invalid model problems = %+v", problems)
	}
	if docs := renderedIfValid(m, problems); len(docs) != 0 {
		t.Fatalf("invalid docs = %+v", docs)
	}
}

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceChecksAndProblemHelpers(t *testing.T) {
	repo := t.TempDir()
	source := filepath.Join(repo, "source.go")
	if err := os.WriteFile(source, []byte("package main\nconst marker = true\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	ok, err := sourceCheckPasses(repo, sourceCheck{Name: "marker", File: "source.go", Contains: "marker"})
	if err != nil || !ok {
		t.Fatalf("source check = %v, %v", ok, err)
	}
	ok, err = sourceCheckPasses(repo, sourceCheck{Name: "missing", File: "source.go", Contains: "nope"})
	if err == nil || ok || !strings.Contains(err.Error(), "missing anchor") {
		t.Fatalf("missing anchor = %v, %v", ok, err)
	}
	if err := problemError([]string{"one", "two"}); err.Error() != "one\ntwo" {
		t.Fatalf("problem error = %q", err)
	}
	if statusFor(nil) != "verified" || statusFor([]string{"x"}) != "failed" {
		t.Fatalf("unexpected status mapping")
	}
}

func TestRenderedIfValidAndManifestValidation(t *testing.T) {
	m := testManifest()
	if docs := renderedIfValid(m, []string{"bad"}); len(docs) != 0 {
		t.Fatalf("invalid docs = %+v", docs)
	}
	if docs := renderedIfValid(m, nil); len(docs) != 4 {
		t.Fatalf("valid docs = %d", len(docs))
	}
	repo := t.TempDir()
	if err := writeText(repoPath(repo, m.Workflow), "workflow"); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	problems, checks := validateManifest(repo, m)
	if len(problems) != 0 || len(checks) != 0 {
		t.Fatalf("validate manifest problems=%+v checks=%+v", problems, checks)
	}
	m.Parts = nil
	problems, _ = validateManifest(repo, m)
	if !hasProblem(problems, "parts") {
		t.Fatalf("manifest problems = %+v", problems)
	}
}

func TestReadJSONRejectsUnknownAndTrailing(t *testing.T) {
	dir := t.TempDir()
	unknown := filepath.Join(dir, "unknown.json")
	if err := os.WriteFile(unknown, []byte(`{"unknown":true}`), 0o644); err != nil {
		t.Fatalf("write unknown: %v", err)
	}
	var m manifest
	if err := readJSON(unknown, &m); err == nil {
		t.Fatal("readJSON accepted unknown field")
	}
	trailing := filepath.Join(dir, "trailing.json")
	if err := os.WriteFile(trailing, []byte(`{"schema_version":"x"} {}`), 0o644); err != nil {
		t.Fatalf("write trailing: %v", err)
	}
	if err := readJSON(trailing, &m); err == nil {
		t.Fatal("readJSON accepted trailing JSON")
	}
}

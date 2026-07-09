package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateAbsentFindsForbiddenTokens(t *testing.T) {
	dir := t.TempDir()
	scope := filepath.Join(dir, "pkg")
	if err := os.MkdirAll(scope, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeText(filepath.Join(scope, "hit.go"), "package pkg\nconst flag = \"forbidden\""); err != nil {
		t.Fatal(err)
	}
	if err := writeText(filepath.Join(scope, "note.txt"), "forbidden"); err != nil {
		t.Fatal(err)
	}
	problems, evidence := validateAbsent(dir, []AbsentSurface{{
		Name: "approval bypass", Scope: []string{"pkg"}, Tokens: []string{"forbidden"},
	}})
	assertToolUseProblem(t, problems, "forbidden token in pkg: forbidden")
	if len(evidence) != 1 || evidence[0].OK {
		t.Fatalf("expected failed absent evidence, got %#v", evidence)
	}
}

func TestValidateAbsentReportsInvalidAndMissingScopes(t *testing.T) {
	dir := t.TempDir()
	problems, evidence := validateAbsent(dir, []AbsentSurface{
		{Name: "", Scope: nil, Tokens: nil},
		{Name: "missing", Scope: []string{"missing"}, Tokens: []string{"token"}},
	})
	for _, want := range []string{"invalid absent surface", "missing"} {
		assertToolUseProblem(t, problems, want)
	}
	if len(evidence) != 1 || evidence[0].OK {
		t.Fatalf("expected missing-scope evidence, got %#v", evidence)
	}
}

func TestScopeContainsFileAndDirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "target.go")
	if err := writeText(file, "package main\nconst value = \"needle\""); err != nil {
		t.Fatal(err)
	}
	found, err := scopeContains(file, "needle")
	if err != nil || !found {
		t.Fatalf("file scan found=%v err=%v", found, err)
	}
	found, err = scopeContains(dir, "absent")
	if err != nil || found {
		t.Fatalf("dir scan found=%v err=%v", found, err)
	}
}

package main

import (
	"encoding/json"
	"go/ast"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesFailedEvidenceForDocDrift(t *testing.T) {
	repo := t.TempDir()
	writeFixtureRepo(t, repo)
	writeFile(t, repo, "docs/20-domain/provider-runtime/adapter-draft-fields/run-lifecycle.md", "stale")
	out := filepath.Join(repo, "evidence.json")
	err := run(t.Context(), options{Repo: repo, Manifest: defaultManifest, CheckDoc: true, EvidenceOut: out})
	if err == nil || !strings.Contains(err.Error(), "generated doc drift") {
		t.Fatalf("expected doc drift error, got %v", err)
	}
	var ev Evidence
	data, readErr := os.ReadFile(out)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if err := json.Unmarshal(data, &ev); err != nil {
		t.Fatal(err)
	}
	if ev.Status != "failed" || len(ev.ProblemSummaries) == 0 {
		t.Fatalf("evidence=%+v", ev)
	}
}

func TestUnsafePathsAndWriteJSONErrors(t *testing.T) {
	repo := t.TempDir()
	if _, err := cleanRepoPath(repo, "../bad.json"); err == nil {
		t.Fatal("expected unsafe path rejection")
	}
	if _, err := loadManifest(repo, "../bad.json"); err == nil {
		t.Fatal("expected unsafe manifest path rejection")
	}
	err := writeJSON(repo, map[string]string{"status": "ok"})
	if err == nil || !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("expected write error, got %v", err)
	}
}

func TestASTHelpersIgnoreNonMatchingNodes(t *testing.T) {
	if statusTypeName(&ast.SelectorExpr{}) != "" {
		t.Fatal("selector type should not be a status type name")
	}
	if stringLiteral(&ast.Ident{Name: "notLiteral"}) != "" {
		t.Fatal("non-literal should return empty string")
	}
	if returnEventType(&ast.CaseClause{Body: []ast.Stmt{&ast.ExprStmt{}}}) != "" {
		t.Fatal("case without return should not resolve event")
	}
	if assignedSelector(&ast.ExprStmt{}) != "" {
		t.Fatal("non-assignment should not resolve selector")
	}
	if selectorName(&ast.Ident{Name: "x"}) != "" {
		t.Fatal("non-selector should not resolve selector name")
	}
}

package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateReportsSourceAbsentAndReferenceProblems(t *testing.T) {
	dir := t.TempDir()
	mustNativeWrite(t, filepath.Join(dir, "src.go"), "secret token")
	m := Manifest{
		SchemaVersion: "v", ID: "x", Title: "X", GeneratedDoc: "doc.md",
		Workflow: "wf.yml", EvidenceArtifact: "e", Purpose: "p",
		Facts:        []Fact{{Name: "fact", Summary: "s", SourceChecks: []string{"missing"}}},
		SourceChecks: []SourceCheck{{Name: "src", File: "src.go", Contains: "secret"}},
		AbsentSurfaces: []AbsentSurface{
			{Name: "no-secret", Scope: []string{"src.go"}, Tokens: []string{"secret"}},
		},
	}
	problems, sources, absent := validate(dir, m)
	if len(sources) != 1 || !sources[0].OK {
		t.Fatalf("source evidence = %#v", sources)
	}
	if len(absent) != 1 || absent[0].OK {
		t.Fatalf("absent evidence = %#v", absent)
	}
	assertNativeProblem(t, problems, "unknown source check missing")
	assertNativeProblem(t, problems, "forbidden token in src.go: secret")
}

func TestValidateReportsInvalidManifestShape(t *testing.T) {
	problems, _, _ := validate(t.TempDir(), Manifest{SourceChecks: []SourceCheck{{}}})
	assertNativeProblem(t, problems, "missing schema_version")
	assertNativeProblem(t, problems, "manifest needs facts and source checks")
	assertNativeProblem(t, problems, "source check with empty name")
}

func assertNativeProblem(t *testing.T, problems []problem, needle string) {
	t.Helper()
	for _, p := range problems {
		if strings.Contains(p.Message, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}

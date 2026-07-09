package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateReportsSourceAbsentAndReferenceProblems(t *testing.T) {
	dir := t.TempDir()
	mustLocalDaemonWrite(t, filepath.Join(dir, "src.go"), "secret token")
	m := localDaemonManifest()
	m.Facts = []Fact{{Name: "fact", Summary: "s", SourceChecks: []string{"missing"}}}
	m.SourceChecks = []SourceCheck{{Name: "src", File: "src.go", Contains: "secret"}}
	m.AbsentSurfaces = []AbsentSurface{
		{Name: "no-secret", Scope: []string{"src.go"}, Tokens: []string{"secret"}},
	}
	problems, sources, absent := validate(dir, m)
	if len(sources) != 1 || !sources[0].OK {
		t.Fatalf("source evidence = %#v", sources)
	}
	if len(absent) != 1 || absent[0].OK {
		t.Fatalf("absent evidence = %#v", absent)
	}
	assertLocalDaemonProblem(t, problems, "unknown source check missing")
	assertLocalDaemonProblem(t, problems, "forbidden token in src.go: secret")
}

func TestValidateReportsInvalidManifestShape(t *testing.T) {
	problems, _, _ := validate(t.TempDir(), Manifest{SourceChecks: []SourceCheck{{}}})
	assertLocalDaemonProblem(t, problems, "missing schema_version")
	assertLocalDaemonProblem(t, problems, "manifest needs facts and source checks")
	assertLocalDaemonProblem(t, problems, "source check with empty name")
	if err := problemError(problems); err == nil || !strings.Contains(err.Error(), "local daemon contract evidence failed") {
		t.Fatalf("expected formatted problem error, got %v", err)
	}
}

func assertLocalDaemonProblem(t *testing.T, problems []problem, needle string) {
	t.Helper()
	for _, p := range problems {
		if strings.Contains(p.Message, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}

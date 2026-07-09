package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateHelpersProduceEvidence(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.go")
	if err := os.WriteFile(source, []byte("package x\nconst marker = true"), 0o644); err != nil {
		t.Fatal(err)
	}
	checks := []SourceCheck{{Name: "marker", File: "source.go", Contains: "marker"}}
	problems, evidence := validateSources(dir, checks)
	if len(problems) != 0 || len(evidence) != 1 || !evidence[0].OK {
		t.Fatalf("source problems=%+v evidence=%+v", problems, evidence)
	}
	missingProblems, _ := validateSources(dir, []SourceCheck{{Name: "bad", File: "source.go", Contains: "nope"}})
	if len(missingProblems) != 1 || !strings.Contains(missingProblems[0].Message, "missing expected") {
		t.Fatalf("missing problems=%+v", missingProblems)
	}
	absentProblems, absent := validateAbsent(dir, []AbsentSurface{{
		Name: "no-secret", Scope: []string{"source.go"}, Tokens: []string{"secret"},
	}})
	if len(absentProblems) != 0 || len(absent) != 1 || !absent[0].OK {
		t.Fatalf("absent problems=%+v evidence=%+v", absentProblems, absent)
	}
	found, err := scopeContains(dir, "marker")
	if err != nil || !found {
		t.Fatalf("scopeContains dir found=%v err=%v", found, err)
	}
}

func TestValidateRefsReportsUnknownAndDuplicate(t *testing.T) {
	unknown := validateRefs(Manifest{Facts: []Fact{{SourceChecks: []string{"missing"}}}})
	if len(unknown) != 1 || !strings.Contains(unknown[0].Message, "unknown source check") {
		t.Fatalf("unknown = %+v", unknown)
	}
	dupe := validateRefs(Manifest{SourceChecks: []SourceCheck{{Name: "same"}, {Name: "same"}}})
	if len(dupe) != 1 || !strings.Contains(dupe[0].Message, "duplicate source check") {
		t.Fatalf("dupe = %+v", dupe)
	}
}

func TestBuildEvidenceAndProblemError(t *testing.T) {
	ev := buildEvidence(
		Manifest{
			ID:            "full-access",
			SchemaVersion: "v1",
			GeneratedDoc:  "doc.md",
			Workflow:      "ci.yml",
			Assertions:    []string{"assert"},
		},
		[]problem{{Message: "bad"}},
		[]SourceCheckEvidence{{Name: "source", OK: true}},
		[]AbsentEvidence{{Name: "surface", OK: true}},
	)
	if ev.ID != "full-access" || len(ev.Problems) != 1 || len(ev.SourceChecks) != 1 {
		t.Fatalf("evidence = %+v", ev)
	}
	err := problemError(ev.Problems)
	if err == nil || !strings.Contains(err.Error(), "1 full-access harness evidence problem") {
		t.Fatalf("problem error = %v", err)
	}
}

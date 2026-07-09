package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSourcesEvidenceAndProblemError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "source.go"), []byte("package x\nmarker"), 0o644); err != nil {
		t.Fatal(err)
	}
	problems, evidence := validateSources(dir, []SourceCheck{{Name: "marker", File: "source.go", Contains: "marker"}})
	if len(problems) != 0 || len(evidence) != 1 || !evidence[0].OK {
		t.Fatalf("problems=%+v evidence=%+v", problems, evidence)
	}
	missing, _ := validateSources(dir, []SourceCheck{{Name: "missing", File: "source.go", Contains: "nope"}})
	if len(missing) != 1 || !strings.Contains(problemError(missing).Error(), "missing expected") {
		t.Fatalf("missing=%+v", missing)
	}
}

func TestBuildEvidenceAndShapeValidation(t *testing.T) {
	policy := PolicySnapshot{Surfaces: []SurfaceSnapshot{
		{ID: serverFacingSurfaceID},
		{ID: providerStatusSurfaceID},
	}}
	problems, shapes := validateShapes(policy)
	if len(problems) == 0 || len(shapes) != 3 {
		t.Fatalf("shape problems=%+v checks=%+v", problems, shapes)
	}
	ev := buildEvidence(
		Manifest{ID: "privacy", SchemaVersion: "v1", GeneratedDoc: "doc.md", Workflow: "ci"},
		policy,
		problems,
		[]SourceCheckEvidence{{Name: "source", OK: true}},
		shapes,
	)
	if ev.ID != "privacy" || len(ev.ShapeChecks) != 3 || len(ev.Problems) == 0 {
		t.Fatalf("evidence = %+v", ev)
	}
	if _, ok := findSurface(policy, "missing"); ok {
		t.Fatal("missing surface should not be found")
	}
}

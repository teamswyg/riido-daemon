package main

import (
	"strings"
	"testing"
)

func TestValidateManifestReportsRequiredSemanticShapes(t *testing.T) {
	m := manifest{
		SchemaVersion: "bad",
		Contexts: []contextRow{
			{ID: "same"},
			{ID: "same", Context: "C", Owner: "O"},
		},
	}
	problems := validateManifest(m)
	for _, want := range []string{
		"unexpected schema_version",
		"id, generated_doc, workflow, and evidence_artifact are required",
		"focused sections and contexts are required",
		"context id, context, and owner are required",
		"duplicate context id same",
		"acl, dependency, and change fragments must not be empty",
		"figma sections must not be empty",
		"split repo rules must not be empty",
	} {
		assertContextMapProblem(t, problems, want)
	}
}

func TestValidateFigmaSectionsRequiresAllBoundaryFacts(t *testing.T) {
	problems := validateFigmaSections([]figmaSection{{Refs: []string{"figma"}}})
	assertContextMapProblem(t, problems, "figma sections require refs, daemon_scope, and not_owned facts")
}

func TestLoadFragmentsReportsMissingFragment(t *testing.T) {
	m := manifest{Fragments: map[string]string{}}
	err := loadFragments(t.TempDir(), "manifest.json", &m)
	if err == nil || !strings.Contains(err.Error(), `missing fragment`) {
		t.Fatalf("expected missing fragment error, got %v", err)
	}
}

func assertContextMapProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}

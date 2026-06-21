package main

import (
	"path/filepath"
	"testing"
)

func TestGeneratedOriginWorkflowCoverageFindsWorkflowReferences(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, ".github/workflows/docs.yml", "run: go run ./tools/example -check-doc\n")
	origins := []generatedOrigin{{Generator: "go run ./tools/example -write-doc", Count: 2}}

	got := scanGeneratedOriginWorkflowCoverage(root, origins)
	if got.CoveredCount != 1 || got.MissingCount != 0 {
		t.Fatalf("coverage = %#v", got)
	}
}

func TestGeneratedOriginWorkflowProblemsReportMissingTool(t *testing.T) {
	root := t.TempDir()
	origins := []generatedOrigin{{Generator: "go run ./tools/missing -write-doc", Count: 1}}

	problems := generatedOriginWorkflowProblems(root, origins)
	if len(problems) != 1 {
		t.Fatalf("problems = %#v", problems)
	}
}

func TestGeneratedToolPathIgnoresFixtureGenerators(t *testing.T) {
	if _, ok := generatedToolPath("fixture"); ok {
		t.Fatal("fixture generator should not require workflow coverage")
	}
}

func TestIsWorkflowFile(t *testing.T) {
	if !isWorkflowFile(filepath.Join(".github", "workflows", "ci.yml")) {
		t.Fatal("expected yml workflow file")
	}
}

package main

import (
	"strings"
	"testing"
)

func TestValidateManifestRejectsWrongSchema(t *testing.T) {
	manifest := validManifest()
	manifest.SchemaVersion = "riido-executable-knowledge-coverage.v1"
	problems := validateManifest(manifest)
	if len(problems) == 0 || !strings.Contains(problems[0], "schema_version must be") {
		t.Fatalf("problems = %v", problems)
	}
}

func TestValidateManifestRequiresWorkflowRoot(t *testing.T) {
	manifest := validManifest()
	manifest.WorkflowRoot = ""
	problems := validateManifest(manifest)
	if len(problems) == 0 || !strings.Contains(strings.Join(problems, "\n"), "workflow_root is required") {
		t.Fatalf("problems = %v", problems)
	}
}

func validManifest() manifest {
	return manifest{
		SchemaVersion:    manifestSchema,
		ID:               "workflow",
		Title:            "Workflow Evidence",
		GeneratedDoc:     "docs/workflow.md",
		Workflow:         ".github/workflows/workflow-evidence.yml",
		EvidenceArtifact: "workflow-evidence",
		WorkflowRoot:     ".github/workflows",
		Assertions:       []string{"workflows publish evidence"},
		Loop: evidenceLoop{
			Observation:   "observe",
			Hypothesis:    "hypothesis",
			Execute:       "execute",
			Evaluate:      "evaluate",
			Retrospective: "retrospective",
		},
	}
}

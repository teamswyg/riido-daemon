package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateManifestRejectsInvalidShapes(t *testing.T) {
	valid := manifest{
		SchemaVersion: manifestSchema,
		ID:            "m",
		LoopSource:    "loop",
		Workflows: []workflowSpec{{
			ID: "w", Workflow: ".github/workflows/w.yml", EvidenceArtifact: "e", RequiredFragments: []string{"go test"},
		}},
	}
	cases := []struct {
		name string
		m    manifest
		want string
	}{
		{name: "schema", m: manifest{}, want: "schema_version"},
		{name: "required", m: manifest{SchemaVersion: manifestSchema}, want: "id, loop_source"},
		{name: "workflow fields", m: manifest{SchemaVersion: manifestSchema, ID: "m", LoopSource: "l", Workflows: []workflowSpec{{ID: "w"}}}, want: "workflow id"},
		{name: "fragments", m: manifest{SchemaVersion: manifestSchema, ID: "m", LoopSource: "l", Workflows: []workflowSpec{{ID: "w", Workflow: "w.yml", EvidenceArtifact: "e"}}}, want: "no required fragments"},
		{name: "duplicate", m: duplicateWorkflowManifest(valid), want: "duplicate workflow"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateManifest(tc.m)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}

func TestFindWorkflowNormalizesPathsAndReportsMissing(t *testing.T) {
	m := manifest{Workflows: []workflowSpec{{ID: "w", Workflow: ".github/workflows/w.yml"}}}
	got, err := findWorkflow(m, filepath.FromSlash(".github/workflows/w.yml"))
	if err != nil || got.ID != "w" {
		t.Fatalf("find workflow=%+v err=%v", got, err)
	}
	if _, err := findWorkflow(m, "missing.yml"); err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("expected missing workflow error, got %v", err)
	}
}

func TestRunFailsWhenWorkflowFragmentsAreMissing(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")
	workflowPath := filepath.Join(dir, "w.yml")
	mustWrite(t, manifestPath, `{"schema_version":"riido-daemon-package-workflow-evidence.v1","id":"m","loop_source":"loop","workflows":[{"id":"w","workflow":"`+filepath.ToSlash(workflowPath)+`","evidence_artifact":"e","required_fragments":["go test"]}]}`)
	mustWrite(t, workflowPath, "run: echo no tests\n")
	err := run([]string{"-manifest", manifestPath, "-workflow", workflowPath, "-evidence-out", filepath.Join(dir, "e.json")})
	if err == nil || !strings.Contains(err.Error(), "evidence status failed") {
		t.Fatalf("expected failed evidence status, got %v", err)
	}
}

func duplicateWorkflowManifest(m manifest) manifest {
	m.Workflows = append(m.Workflows, m.Workflows[0])
	return m
}

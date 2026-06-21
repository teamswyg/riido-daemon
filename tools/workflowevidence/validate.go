package main

import "fmt"

func validateManifest(m manifest) []string {
	var problems []string
	if m.SchemaVersion != manifestSchema {
		problems = append(problems, fmt.Sprintf("schema_version must be %s", manifestSchema))
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" ||
		m.Workflow == "" || m.EvidenceArtifact == "" {
		problems = append(problems, "id, title, generated_doc, workflow, and evidence_artifact are required")
	}
	if m.WorkflowRoot == "" {
		problems = append(problems, "workflow_root is required")
	}
	if len(m.Assertions) == 0 {
		problems = append(problems, "assertions must not be empty")
	}
	return append(problems, validateLoop(m.Loop)...)
}

func validateLoop(item evidenceLoop) []string {
	var problems []string
	if item.Observation == "" || item.Hypothesis == "" || item.Execute == "" ||
		item.Evaluate == "" || item.Retrospective == "" {
		problems = append(problems, "loop must include observation, hypothesis, execute, evaluate, and retrospective")
	}
	return problems
}

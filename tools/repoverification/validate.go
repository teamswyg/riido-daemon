package main

import (
	"fmt"
	"os"
)

func validateManifest(root string, m manifest) []string {
	var problems []string
	if m.SchemaVersion != "riido-repo-verification.v1" {
		problems = append(problems, "schema_version must be riido-repo-verification.v1")
	}
	for _, value := range []string{m.ID, m.Title, m.GeneratedDoc, m.Workflow, m.EvidenceArtifact} {
		if value == "" {
			problems = append(problems, "id, title, generated_doc, workflow, and evidence_artifact are required")
		}
	}
	if _, err := os.Stat(resolvePath(root, m.Workflow)); err != nil {
		problems = append(problems, fmt.Sprintf("missing workflow %q", m.Workflow))
	}
	problems = append(problems, validateCommands(m.Commands)...)
	if len(m.Assertions) == 0 {
		problems = append(problems, "assertions must not be empty")
	}
	return problems
}

func validateCommands(commands []commandSpec) []string {
	var problems []string
	if len(commands) == 0 {
		return []string{"commands must not be empty"}
	}
	seen := map[string]bool{}
	for _, command := range commands {
		problems = append(problems, validateCommand(command, seen)...)
	}
	return problems
}

package main

import (
	"fmt"
	"os"
)

func validateManifest(root string, m manifest) []string {
	var problems []string
	if m.SchemaVersion != "riido-executable-knowledge-coverage.v1" {
		problems = append(problems, "schema_version must be riido-executable-knowledge-coverage.v1")
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		problems = append(problems, "id, title, generated_doc, workflow, and evidence_artifact are required")
	}
	if len(m.ScanRoots) == 0 || len(m.Assertions) == 0 {
		problems = append(problems, "scan_roots and assertions must not be empty")
	}
	for _, path := range append([]string{m.Workflow}, m.ScanRoots...) {
		if _, err := os.Stat(resolvePath(root, path)); err != nil {
			problems = append(problems, fmt.Sprintf("missing path %q", path))
		}
	}
	return append(problems, validateManualGroups(m)...)
}

func validateManualGroups(m manifest) []string {
	var problems []string
	seenGroups := map[string]bool{}
	seenPaths := map[string]string{}
	for _, group := range m.ManualGroups {
		if group.ID == "" || group.Owner == "" || group.Reason == "" || group.NextArtifact == "" {
			problems = append(problems, "manual group id, owner, reason, and next_artifact are required")
		}
		if seenGroups[group.ID] {
			problems = append(problems, fmt.Sprintf("duplicate manual group %q", group.ID))
		}
		seenGroups[group.ID] = true
		if len(group.Paths) == 0 {
			problems = append(problems, fmt.Sprintf("manual group %q has no paths", group.ID))
		}
		problems = append(problems, validateManualGroupPaths(group, seenPaths)...)
	}
	return problems
}

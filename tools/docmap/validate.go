package main

import (
	"fmt"
	"os"
)

func validateManifest(root string, m manifest) []string {
	var problems []string
	if m.SchemaVersion != "riido-doc-map.v1" {
		problems = append(problems, "schema_version must be riido-doc-map.v1")
	}
	for _, value := range []string{m.ID, m.Title, m.GeneratedDocs.Readme, m.GeneratedDocs.DocumentMap, m.EvidenceArtifact, m.Intro} {
		if value == "" {
			problems = append(problems, "id, title, generated docs, evidence_artifact, and intro are required")
		}
	}
	problems = append(problems, requireNonEmpty(m)...)
	problems = append(problems, validateUniqueTopics(m.Decisions)...)
	problems = append(problems, validateDocRefs(root, m)...)
	return problems
}

func requireNonEmpty(m manifest) []string {
	var problems []string
	if len(m.ReadOrder) == 0 || len(m.Decisions) == 0 || len(m.Repos) == 0 || len(m.Rules) == 0 {
		problems = append(problems, "read_order, decisions, repos, and rules must not be empty")
	}
	return problems
}

func validateDocRefs(root string, m manifest) []string {
	var problems []string
	for _, doc := range generatedDocPaths(m) {
		problems = append(problems, requireDoc(root, doc)...)
	}
	for _, row := range m.ReadOrder {
		problems = append(problems, validateReadEntry(root, row)...)
	}
	for _, row := range m.Decisions {
		problems = append(problems, validateDecision(root, row)...)
	}
	return problems
}

func requireDoc(root, doc string) []string {
	if _, err := os.Stat(resolvePath(root, doc)); err != nil {
		return []string{fmt.Sprintf("missing doc %q", doc)}
	}
	return nil
}

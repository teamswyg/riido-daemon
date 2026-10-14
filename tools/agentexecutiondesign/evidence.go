package main

import "sort"

func buildEvidence(m model, docs map[string]string, problems []string) result {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	paths := make([]string, 0, len(docs))
	for path := range docs {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return result{
		SchemaVersion:    "riido-agent-execution-design-docs-result.v1",
		ID:               m.Manifest.ID,
		Status:           status,
		GeneratedDocs:    paths,
		EvidenceItems:    len(m.Items),
		Remaining:        len(m.Boundaries),
		ProblemSummaries: problems,
		EvidenceArtifact: m.Manifest.EvidenceArtifact,
	}
}

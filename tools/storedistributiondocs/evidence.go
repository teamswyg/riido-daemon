package main

import "sort"

func buildEvidence(m manifest, c contract, docs map[string]string, problems []string) evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	var channels []string
	for _, item := range c.Channels {
		channels = append(channels, item.ID)
	}
	var paths []string
	for path := range docs {
		paths = append(paths, path)
	}
	sort.Strings(channels)
	sort.Strings(paths)
	return evidence{
		SchemaVersion:    "riido-store-distribution-docs-result.v1",
		ID:               m.ID,
		Status:           status,
		GeneratedDocs:    paths,
		Channels:         channels,
		ProblemSummaries: problems,
		EvidenceArtifact: m.EvidenceArtifact,
	}
}

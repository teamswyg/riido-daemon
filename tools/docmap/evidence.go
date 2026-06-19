package main

import "time"

func buildEvidence(m manifest) evidenceFile {
	return evidenceFile{
		SchemaVersion:  "riido-doc-map-evidence-result.v1",
		ID:             m.ID,
		ObservedAt:     time.Now().UTC().Format(time.RFC3339),
		Status:         "verified",
		GeneratedDocs:  generatedDocPaths(m),
		ReadOrderCount: len(m.ReadOrder),
		DecisionCount:  len(m.Decisions),
		RepoCount:      len(m.Repos),
		RuleCount:      len(m.Rules),
		Assertions: []string{
			"document map reader docs are generated from one manifest",
			"every mapped document path exists",
			"decision topics are unique",
		},
	}
}

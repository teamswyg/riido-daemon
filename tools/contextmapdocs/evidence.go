package main

func buildEvidence(m manifest, docs map[string]string, problems []string) evidence {
	return evidence{
		SchemaVersion:      "riido-context-map-docs-result.v1",
		ID:                 m.ID,
		Status:             statusFor(problems),
		GeneratedDocs:      generatedDocPaths(m),
		ContextCount:       len(m.Contexts),
		ACLCount:           len(m.ACL.Rows),
		FigmaBoundaryCount: len(m.FigmaDaemon.Sections) + len(m.FigmaOnboarding.Sections),
		ProblemSummaries:   problems,
		EvidenceArtifact:   m.EvidenceArtifact,
	}
}

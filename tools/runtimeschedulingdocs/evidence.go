package main

func buildEvidence(m manifest, checks []sourceCheckResult, problems []string) evidence {
	return evidence{
		SchemaVersion:    "riido-runtime-scheduling-docs-result.v1",
		ID:               m.ID,
		Status:           statusFor(problems),
		GeneratedDocs:    generatedDocPaths(m),
		SourceChecks:     checks,
		AssertionCount:   len(m.Assertions),
		InvariantCount:   len(m.Core.Invariants),
		ProblemSummaries: problems,
		EvidenceArtifact: m.EvidenceArtifact,
	}
}

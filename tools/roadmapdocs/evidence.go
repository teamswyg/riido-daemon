package main

func buildEvidence(m manifest, checks []sourceCheckResult, problems []string) evidence {
	return evidence{
		SchemaVersion:    "riido-roadmap-docs-result.v1",
		ID:               m.ID,
		Status:           statusFor(problems),
		GeneratedDoc:     m.GeneratedDoc,
		QuestionCount:    len(m.Questions),
		SourceChecks:     checks,
		AssertionCount:   len(m.Assertions),
		ProblemSummaries: problems,
		EvidenceArtifact: m.EvidenceArtifact,
	}
}

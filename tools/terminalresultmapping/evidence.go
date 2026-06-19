package main

func buildEvidence(
	manifest Manifest,
	problems []problem,
	mappings []MappingCheck,
	coverage []CoverageCheck,
	defaults []DefaultCheck,
) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion:    "riido-terminal-result-mapping-result.v1",
		ID:               manifest.ID,
		Status:           status,
		MappingChecks:    mappings,
		CoverageChecks:   coverage,
		DefaultChecks:    defaults,
		Assertions:       manifest.Assertions,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: manifest.EvidenceArtifact,
	}
}

package main

func buildEvidence(
	manifest Manifest,
	problems []problem,
	sources []SourceResult,
	absent []AbsentCheck,
) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion:    "riido-session-lifecycle-result.v1",
		ID:               manifest.ID,
		Status:           status,
		SourceChecks:     sources,
		AbsentChecks:     absent,
		Assertions:       manifest.Assertions,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: manifest.EvidenceArtifact,
	}
}

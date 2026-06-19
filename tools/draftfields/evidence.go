package main

func buildEvidence(
	manifest Manifest,
	problems []problem,
	allowed []AllowedCheck,
	forbidden []ForbiddenCheck,
) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion:    "riido-draft-field-surface-result.v1",
		ID:               manifest.ID,
		Status:           status,
		AllowedChecks:    allowed,
		ForbiddenChecks:  forbidden,
		Assertions:       manifest.Assertions,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: manifest.EvidenceArtifact,
	}
}

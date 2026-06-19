package main

func buildEvidence(
	manifest Manifest,
	problems []problem,
	levels []LevelCheck,
	timeouts []TimeoutCheck,
	consumers []ConsumerCheck,
) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion:    "riido-shutdown-authority-result.v1",
		ID:               manifest.ID,
		Status:           status,
		LevelChecks:      levels,
		TimeoutChecks:    timeouts,
		ConsumerChecks:   consumers,
		Assertions:       manifest.Assertions,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: manifest.EvidenceArtifact,
	}
}

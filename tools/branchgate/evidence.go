package main

func buildEvidence(
	manifest Manifest,
	problems []problem,
	scripts []ScriptCheck,
	examples []ExampleCheck,
) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion:    "riido-work-branch-gate-result.v1",
		ID:               manifest.ID,
		Status:           status,
		ScriptChecks:     scripts,
		ExampleChecks:    examples,
		Assertions:       manifest.Assertions,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: manifest.EvidenceArtifact,
	}
}

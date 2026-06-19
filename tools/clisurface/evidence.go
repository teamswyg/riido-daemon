package main

func buildEvidence(
	manifest Manifest,
	problems []problem,
	sources []CheckResult,
	forbidden []CheckResult,
	behaviors []CheckResult,
) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion:    "riido-cli-surface-evidence-result.v1",
		ID:               manifest.ID,
		Status:           status,
		SourceChecks:     sources,
		ForbiddenChecks:  forbidden,
		BehaviorChecks:   behaviors,
		CommandGroups:    commandGroupNames(manifest.CommandGroups),
		Providers:        manifest.Providers,
		Assertions:       manifest.Assertions,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: manifest.EvidenceArtifact,
	}
}

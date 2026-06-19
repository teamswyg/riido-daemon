package main

func buildEvidence(
	manifest Manifest,
	problems []problem,
	envConsts []CheckResult,
	anchors []CheckResult,
	sources []CheckResult,
) Evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return Evidence{
		SchemaVersion:    "riido-config-reference-evidence-result.v1",
		ID:               manifest.ID,
		Status:           status,
		EnvConstChecks:   envConsts,
		AnchorChecks:     anchors,
		SourceChecks:     sources,
		EnvVarCount:      len(manifest.DaemonEnvVars),
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: manifest.EvidenceArtifact,
	}
}

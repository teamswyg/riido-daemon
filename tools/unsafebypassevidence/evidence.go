package main

func buildEvidence(
	manifest Manifest,
	problems []problem,
	sources []SourceEvidence,
	policies []PolicyEvidence,
	codexArgs []CodexArgEvidence,
) Evidence {
	return Evidence{
		ID:             manifest.ID,
		Artifact:       manifest.EvidenceArtifact,
		Problems:       problemMessages(problems),
		SourceChecks:   sources,
		PolicyChecks:   policies,
		CodexArgChecks: codexArgs,
		Assertions:     manifest.Assertions,
	}
}

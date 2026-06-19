package main

func buildEvidence(
	manifest Manifest,
	policy PolicySnapshot,
	problems []problem,
	sources []SourceCheckEvidence,
	shapes []ShapeCheck,
) Evidence {
	return Evidence{
		ID:             manifest.ID,
		SchemaVersion:  manifest.SchemaVersion,
		GeneratedDoc:   manifest.GeneratedDoc,
		Workflow:       manifest.Workflow,
		PolicyArtifact: manifest.PolicyArtifact,
		Problems:       problems,
		Policy:         policy,
		SourceChecks:   sources,
		ShapeChecks:    shapes,
	}
}

package main

func buildEvidence(
	manifest Manifest,
	problems []problem,
	sources []SourceCheckEvidence,
	absent []AbsentEvidence,
) Evidence {
	return Evidence{
		ID:             manifest.ID,
		SchemaVersion:  manifest.SchemaVersion,
		GeneratedDoc:   manifest.GeneratedDoc,
		Workflow:       manifest.Workflow,
		Problems:       problems,
		SourceChecks:   sources,
		AbsentSurfaces: absent,
		Assertions:     manifest.Assertions,
	}
}

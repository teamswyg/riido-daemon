package main

func buildEvidence(m Manifest, problems []problem, sources []SourceEvidence, absent []AbsentEvidence) Evidence {
	return Evidence{
		ID:             m.ID,
		Artifact:       m.EvidenceArtifact,
		Problems:       problemMessages(problems),
		SourceChecks:   sources,
		AbsentSurfaces: absent,
		Assertions:     m.Assertions,
	}
}

func problemMessages(problems []problem) []string {
	out := make([]string, 0, len(problems))
	for _, p := range problems {
		out = append(out, p.Message)
	}
	return out
}

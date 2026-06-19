package main

type Evidence struct {
	SchemaVersion string            `json:"schema_version"`
	ID            string            `json:"id"`
	ProblemCount  int               `json:"problem_count"`
	Problems      []problem         `json:"problems"`
	Sources       []SourceEvidence  `json:"sources"`
	Surfaces      []SurfaceEvidence `json:"surfaces"`
	Assertions    []string          `json:"assertions"`
}

func buildEvidence(
	m Manifest,
	problems []problem,
	sources []SourceEvidence,
	surfaces []SurfaceEvidence,
) Evidence {
	return Evidence{
		SchemaVersion: "riido-task-requirements-result.v1",
		ID:            m.ID,
		ProblemCount:  len(problems),
		Problems:      problems,
		Sources:       sources,
		Surfaces:      surfaces,
		Assertions:    m.Assertions,
	}
}

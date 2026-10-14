package main

type Evidence struct {
	SchemaVersion string             `json:"schema_version"`
	ID            string             `json:"id"`
	ProblemCount  int                `json:"problem_count"`
	Problems      []problem          `json:"problems"`
	Sources       []SourceEvidence   `json:"sources"`
	Behaviors     []BehaviorEvidence `json:"behaviors"`
	Assertions    []string           `json:"assertions"`
}

func buildEvidence(
	m Manifest,
	problems []problem,
	sources []SourceEvidence,
	behaviors []BehaviorEvidence,
) Evidence {
	return Evidence{
		SchemaVersion: "riido-executable-search-path-result.v1",
		ID:            m.ID,
		ProblemCount:  len(problems),
		Problems:      problems,
		Sources:       sources,
		Behaviors:     behaviors,
		Assertions:    m.Assertions,
	}
}

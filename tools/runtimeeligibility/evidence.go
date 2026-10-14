package main

type Evidence struct {
	SchemaVersion string           `json:"schema_version"`
	ID            string           `json:"id"`
	ProblemCount  int              `json:"problem_count"`
	Problems      []problem        `json:"problems"`
	Sources       []SourceEvidence `json:"sources"`
	Gates         []GateEvidence   `json:"gates"`
	Absent        []AbsentEvidence `json:"absent_scans"`
	Assertions    []string         `json:"assertions"`
}

func buildEvidence(
	m Manifest,
	problems []problem,
	sources []SourceEvidence,
	gates []GateEvidence,
	absent []AbsentEvidence,
) Evidence {
	return Evidence{
		SchemaVersion: "riido-runtime-eligibility-result.v1",
		ID:            m.ID,
		ProblemCount:  len(problems),
		Problems:      problems,
		Sources:       sources,
		Gates:         gates,
		Absent:        absent,
		Assertions:    m.Assertions,
	}
}

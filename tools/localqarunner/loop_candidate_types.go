package main

type runLoopCandidate struct {
	ID             string                `json:"id"`
	SourceScenario string                `json:"source_scenario"`
	Class          string                `json:"class"`
	Reason         string                `json:"reason"`
	NextEvidence   string                `json:"next_evidence"`
	Graph          runLoopCandidateGraph `json:"evidence_graph"`
}

type runLoopCandidateGraph struct {
	Observation string `json:"observation"`
	Hypothesis  string `json:"hypothesis"`
	Change      string `json:"change"`
	Verifier    string `json:"verifier"`
	Evidence    string `json:"evidence"`
	Decision    string `json:"decision"`
	NextLoop    string `json:"next_loop"`
}

package main

type evidenceGapCandidate struct {
	ID             string `json:"id"`
	SourceScenario string `json:"source_scenario"`
	Class          string `json:"class"`
	Reason         string `json:"reason"`
	NextEvidence   string `json:"next_evidence"`
}

type evidenceGapSummary struct {
	Skipped                []string
	FigmaWithoutScreenshot []string
}

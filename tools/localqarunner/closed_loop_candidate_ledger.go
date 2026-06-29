package main

import (
	"encoding/json"
	"os"
)

func loadPreviousCandidates(path string) []closedLoopCandidate {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var previous runEvidence
	if json.Unmarshal(data, &previous) != nil {
		return nil
	}
	return previous.Candidates
}

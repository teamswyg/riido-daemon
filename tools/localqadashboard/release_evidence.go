package main

import "encoding/json"

func releaseEvidenceScenarios(path string) []externalScenario {
	data, ok := readOptional(path)
	if !ok {
		return nil
	}
	var evidence externalEvidenceFile
	if json.Unmarshal(data, &evidence) != nil {
		return nil
	}
	return evidence.Scenarios
}

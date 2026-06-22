package main

import "encoding/json"

type scheduleEvidence struct {
	Status              string `json:"status"`
	CommandHasTokenText bool   `json:"command_has_token_text"`
	S3PrefixConfigured  bool   `json:"s3_prefix_configured"`
}

func scheduleEvidenceScenarios(path string) []externalScenario {
	data, ok := readOptional(path)
	if !ok {
		return nil
	}
	var evidence scheduleEvidence
	if json.Unmarshal(data, &evidence) != nil {
		return nil
	}
	if evidence.Status == "" {
		return nil
	}
	scenario := externalScenario{ID: "local.qa.daily_schedule", Status: evidence.Status}
	if evidence.CommandHasTokenText || !evidence.S3PrefixConfigured {
		scenario.Status = statusFailed
	}
	return []externalScenario{scenario}
}

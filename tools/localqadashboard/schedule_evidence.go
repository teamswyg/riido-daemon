package main

import (
	"encoding/json"
	"fmt"
)

type scheduleEvidence struct {
	Status              string `json:"status"`
	Label               string `json:"label"`
	PlistPath           string `json:"plist_path"`
	StdoutPath          string `json:"stdout_path"`
	StderrPath          string `json:"stderr_path"`
	Hour                int    `json:"hour"`
	Minute              int    `json:"minute"`
	RunAtLoad           bool   `json:"run_at_load"`
	CommandHasTokenText bool   `json:"command_has_token_text"`
	S3PrefixConfigured  bool   `json:"s3_prefix_configured"`
	CommandPreview      string `json:"command_preview"`
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
	scenario := externalScenario{
		ID:             "local.qa.daily_schedule",
		Status:         evidence.Status,
		FailureSummary: scheduleEvidenceDetail(evidence),
	}
	if evidence.CommandHasTokenText || !evidence.S3PrefixConfigured {
		scenario.Status = statusFailed
	}
	return []externalScenario{scenario}
}

func scheduleEvidenceDetail(e scheduleEvidence) string {
	return fmt.Sprintf(
		"label=%s time=%02d:%02d run_at_load=%t plist=%s stdout=%s stderr=%s command=%s",
		e.Label, e.Hour, e.Minute, e.RunAtLoad, e.PlistPath, e.StdoutPath, e.StderrPath, e.CommandPreview,
	)
}

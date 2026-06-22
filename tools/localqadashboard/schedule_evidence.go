package main

import (
	"encoding/json"
	"fmt"
)

type scheduleEvidence struct {
	Status              string          `json:"status"`
	Label               string          `json:"label"`
	Installed           bool            `json:"installed"`
	PlistPath           string          `json:"plist_path"`
	StdoutPath          string          `json:"stdout_path"`
	StderrPath          string          `json:"stderr_path"`
	Hour                int             `json:"hour"`
	Minute              int             `json:"minute"`
	RunAtLoad           bool            `json:"run_at_load"`
	CommandHasTokenText bool            `json:"command_has_token_text"`
	S3PrefixConfigured  bool            `json:"s3_prefix_configured"`
	CommandPreview      string          `json:"command_preview"`
	Launchd             launchdEvidence `json:"launchd"`
}

type launchdEvidence struct {
	Checked         bool   `json:"checked"`
	Loaded          bool   `json:"loaded"`
	State           string `json:"state"`
	Runs            int    `json:"runs"`
	LastExitCode    string `json:"last_exit_code"`
	CalendarTrigger bool   `json:"calendar_trigger"`
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
	if scheduleEvidenceFailed(evidence) {
		scenario.Status = statusFailed
	}
	return []externalScenario{scenario}
}

func scheduleEvidenceFailed(e scheduleEvidence) bool {
	if !e.Installed || e.CommandHasTokenText || !e.S3PrefixConfigured {
		return true
	}
	if !e.Launchd.Checked || !e.Launchd.Loaded || !e.Launchd.CalendarTrigger {
		return true
	}
	return e.Launchd.Runs <= 0 || e.Launchd.LastExitCode != "0"
}

func scheduleEvidenceDetail(e scheduleEvidence) string {
	return fmt.Sprintf(
		"label=%s time=%02d:%02d run_at_load=%t launchd_loaded=%t launchd_state=%s runs=%d last_exit=%s calendar_trigger=%t plist=%s stdout=%s stderr=%s command=%s",
		e.Label, e.Hour, e.Minute, e.RunAtLoad, e.Launchd.Loaded, e.Launchd.State,
		e.Launchd.Runs, e.Launchd.LastExitCode, e.Launchd.CalendarTrigger,
		e.PlistPath, e.StdoutPath, e.StderrPath, e.CommandPreview,
	)
}

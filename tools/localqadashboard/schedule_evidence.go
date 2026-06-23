package main

import "encoding/json"

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
	CoverageEvidence    string          `json:"coverage_evidence"`
	CommandPreview      string          `json:"command_preview"`
	Trigger             triggerEvidence `json:"trigger"`
	Launchd             launchdEvidence `json:"launchd"`
}

type triggerEvidence struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CommonName string `json:"common_name"`
	Cadence    string `json:"cadence"`
	TimeLocal  string `json:"time_local"`
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
		Evidence:       path,
	}
	if scheduleEvidenceFailed(evidence) {
		scenario.Status = statusFailed
	}
	triggerScenario := externalScenario{
		ID:             "local.qa.daily_trigger",
		Status:         evidence.Status,
		FailureSummary: triggerEvidenceDetail(evidence),
		Evidence:       path,
	}
	if scheduleEvidenceFailed(evidence) || evidence.Trigger.ID != "daily-evidence-sweep" {
		triggerScenario.Status = statusFailed
	}
	return []externalScenario{scenario, triggerScenario}
}

func scheduleEvidenceFailed(e scheduleEvidence) bool {
	if !e.Installed || e.CommandHasTokenText || !e.S3PrefixConfigured || e.CoverageEvidence == "" {
		return true
	}
	if !e.Launchd.Checked || !e.Launchd.Loaded || !e.Launchd.CalendarTrigger {
		return true
	}
	if e.Launchd.Runs <= 0 {
		return true
	}
	if e.Launchd.LastExitCode == "0" {
		return false
	}
	return e.Launchd.State != "running"
}

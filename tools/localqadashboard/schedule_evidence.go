package main

import "encoding/json"

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
	return []externalScenario{
		scenario,
		triggerScenario(path, evidence),
		freshnessScenario(path, evidence),
	}
}

func triggerScenario(path string, e scheduleEvidence) externalScenario {
	row := externalScenario{
		ID:             "local.qa.daily_trigger",
		Status:         e.Status,
		FailureSummary: triggerEvidenceDetail(e),
		Evidence:       path,
	}
	if scheduleEvidenceFailed(e) || e.Trigger.ID != "daily-evidence-sweep" {
		row.Status = statusFailed
	}
	return row
}

func freshnessScenario(path string, e scheduleEvidence) externalScenario {
	row := externalScenario{
		ID:             "local.qa.daily_freshness",
		Status:         e.Status,
		FailureSummary: freshnessEvidenceDetail(e),
		Evidence:       path,
	}
	if scheduleEvidenceFailed(e) || freshnessEvidenceFailed(e) {
		row.Status = statusFailed
	}
	return row
}

package main

import (
	"slices"
	"strings"
)

func buildQASchedule(m manifest, source qaScheduleSource) qaScheduleEvidence {
	out := qaScheduleEvidence{
		ID:                      source.ID,
		Cadence:                 source.Cadence,
		FreshnessWindow:         source.FreshnessWindow,
		Entrypoint:              source.Entrypoint,
		EvidenceCount:           len(source.Evidence),
		ClosedLoopStepCount:     len(source.ClosedLoop),
		RejectConditionCount:    len(source.RejectIf),
		LocalQARunEvidenceBound: slices.Contains(source.Evidence, m.LocalQARunEvidence),
		ScheduleEvidenceBound:   containsSuffix(source.Evidence, "local-qa-schedule.json"),
		DashboardHandoffBound:   containsDashboardHandoff(source.Evidence),
		Status:                  statusPassed,
	}
	if !scheduleComplete(out) {
		out.Status = statusPartial
		out.PartialReason = "local QA schedule does not prove run evidence, schedule evidence, and dashboard handoff"
	}
	return out
}

func scheduleComplete(v qaScheduleEvidence) bool {
	return v.ID != "" && v.Cadence == "daily" &&
		strings.Contains(v.Entrypoint, "localqarunner") &&
		v.FreshnessWindow == "24h" &&
		v.LocalQARunEvidenceBound &&
		v.ScheduleEvidenceBound &&
		v.DashboardHandoffBound
}

func containsSuffix(items []string, suffix string) bool {
	for _, item := range items {
		if strings.HasSuffix(item, suffix) {
			return true
		}
	}
	return false
}

func containsDashboardHandoff(items []string) bool {
	for _, item := range items {
		if strings.Contains(item, "dashboard") || strings.Contains(item, ".html") {
			return true
		}
	}
	return false
}

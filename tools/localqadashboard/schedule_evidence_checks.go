package main

import "strings"

func scheduleEvidenceFailed(e scheduleEvidence) bool {
	if !e.Installed || e.CommandHasTokenText || !e.S3PrefixConfigured {
		return true
	}
	if e.CoverageEvidence == "" {
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

func freshnessEvidenceFailed(e scheduleEvidence) bool {
	if !e.Trigger.RefreshesExpiredEvidence || e.Trigger.FreshnessWindow == "" {
		return true
	}
	if !strings.Contains(e.Trigger.ClosedLoop, "expires_at") {
		return true
	}
	return len(e.Trigger.Evidence) == 0
}

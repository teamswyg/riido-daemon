package main

import "fmt"

func scheduleEvidenceDetail(e scheduleEvidence) string {
	return fmt.Sprintf(
		"label=%s time=%02d:%02d run_at_load=%t launchd_loaded=%t launchd_state=%s runs=%d last_exit=%s calendar_trigger=%t coverage=%s plist=%s stdout=%s stderr=%s command=%s",
		e.Label, e.Hour, e.Minute, e.RunAtLoad, e.Launchd.Loaded, e.Launchd.State,
		e.Launchd.Runs, e.Launchd.LastExitCode, e.Launchd.CalendarTrigger,
		e.CoverageEvidence, e.PlistPath, e.StdoutPath, e.StderrPath, e.CommandPreview,
	)
}

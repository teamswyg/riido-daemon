package main

import "testing"

func TestParseLaunchdPrintReadsLiveScheduleFields(t *testing.T) {
	out := `state = not running
runs = 3
last exit code = 0
stream = com.apple.launchd.calendarinterval
`
	got := parseLaunchdPrint(out)
	if got.State != "not running" || got.Runs != 3 || got.LastExitCode != "0" {
		t.Fatalf("launchd=%+v", got)
	}
	if !got.CalendarTrigger {
		t.Fatalf("calendar trigger missing: %+v", got)
	}
}

func TestParseLaunchdPrintPreservesNeverExited(t *testing.T) {
	got := parseLaunchdPrint("last exit code = (never exited)\n")
	if got.LastExitCode != "(never exited)" {
		t.Fatalf("last exit=%q", got.LastExitCode)
	}
}

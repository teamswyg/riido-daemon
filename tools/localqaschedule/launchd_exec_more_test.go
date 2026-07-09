package main

import (
	"strings"
	"testing"
)

func TestInstallLaunchAgentUsesLaunchdCommand(t *testing.T) {
	var calls []string
	restore := fakeLaunchd(t, "0", "", &calls)
	defer restore()
	err := installLaunchAgent(schedulePaths{launchctl: "launchctl", plist: "qa.plist"})
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || !strings.Contains(calls[1], "bootstrap") {
		t.Fatalf("calls = %#v", calls)
	}
}

func TestInstallLaunchAgentReportsBootstrapFailure(t *testing.T) {
	restore := fakeLaunchd(t, "7", "boom", nil)
	defer restore()
	err := installLaunchAgent(schedulePaths{launchctl: "launchctl", plist: "qa.plist"})
	if err == nil || !strings.Contains(err.Error(), "launchctl bootstrap") {
		t.Fatalf("expected bootstrap error, got %v", err)
	}
}

func TestInspectLaunchAgentParsesLiveLaunchdEvidence(t *testing.T) {
	out := "state = running\nruns = 2\nlast exit code = 0\ncom.apple.launchd.calendarinterval\n"
	restore := fakeLaunchd(t, "0", out, nil)
	defer restore()
	got, err := inspectLaunchAgent(schedulePaths{launchctl: "launchctl"}, "io.test")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Checked || !got.Loaded || got.Domain != "gui/501" || !got.CalendarTrigger {
		t.Fatalf("launchd evidence = %+v", got)
	}
}

func TestLaunchdEvidenceForRunRequiresCalendarTrigger(t *testing.T) {
	restore := fakeLaunchd(t, "0", "state = running\n", nil)
	defer restore()
	cfg := testConfig()
	inspect := true
	cfg.inspect = &inspect
	_, err := launchdEvidenceForRun(cfg, schedulePaths{launchctl: "launchctl"})
	if err == nil || !strings.Contains(err.Error(), "calendar trigger missing") {
		t.Fatalf("expected calendar trigger error, got %v", err)
	}
}

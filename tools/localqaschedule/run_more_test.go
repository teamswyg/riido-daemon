package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesPlistAndEvidenceWithoutLaunchd(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()
	evidenceOut := filepath.Join(dir, "evidence", "schedule.json")
	plistPath := filepath.Join(dir, "launchd", "qa.plist")
	repo := filepath.Join(dir, "repo")
	cfg.repo = &repo
	cfg.evidenceOut = &evidenceOut
	cfg.plistPath = &plistPath

	got, err := run(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got != plistPath {
		t.Fatalf("plist path = %q, want %q", got, plistPath)
	}
	plist, err := os.ReadFile(plistPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(plist), "io.test") {
		t.Fatalf("plist missing label: %s", plist)
	}
	var ev scheduleEvidence
	readScheduleJSON(t, evidenceOut, &ev)
	if ev.Installed || ev.PlistPath != plistPath || ev.Status != "passed" {
		t.Fatalf("evidence = %+v", ev)
	}
}

func TestValidateTimeBounds(t *testing.T) {
	for _, tt := range []struct {
		hour   int
		minute int
	}{
		{hour: -1, minute: 0},
		{hour: 24, minute: 0},
		{hour: 0, minute: -1},
		{hour: 0, minute: 60},
	} {
		if err := validateTime(tt.hour, tt.minute); err == nil {
			t.Fatalf("validateTime(%d,%d) = nil", tt.hour, tt.minute)
		}
	}
	if err := validateTime(23, 59); err != nil {
		t.Fatalf("validateTime valid = %v", err)
	}
}

func TestGetenvDefault(t *testing.T) {
	t.Setenv("RIIDO_TEST_ENV", "custom")
	if got := getenvDefault("RIIDO_TEST_ENV", "fallback"); got != "custom" {
		t.Fatalf("getenv existing = %q", got)
	}
	if got := getenvDefault("RIIDO_TEST_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("getenv fallback = %q", got)
	}
}

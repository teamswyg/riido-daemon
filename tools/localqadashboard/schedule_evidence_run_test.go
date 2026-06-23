package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScheduleEvidenceScenariosAcceptsActiveLaunchdRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schedule.json")
	body := `{"status":"passed","installed":true,"s3_prefix_configured":true,"coverage_evidence":"coverage.json","trigger":{"id":"daily-evidence-sweep"},"launchd":{"checked":true,"loaded":true,"state":"running","calendar_trigger":true,"runs":1,"last_exit_code":"(never exited)"}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := scheduleEvidenceScenarios(path)
	if len(got) != 2 || got[0].Status != statusPassed {
		t.Fatalf("scenarios=%+v", got)
	}
}

func TestScheduleEvidenceScenariosRequiresCoverageEvidencePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schedule.json")
	body := `{"status":"passed","installed":true,"s3_prefix_configured":true,"launchd":{"checked":true,"loaded":true,"state":"not running","calendar_trigger":true,"runs":1,"last_exit_code":"0"}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := scheduleEvidenceScenarios(path)
	if len(got) != 2 || got[0].Status != statusFailed {
		t.Fatalf("scenarios=%+v", got)
	}
}

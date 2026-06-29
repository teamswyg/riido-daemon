package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScheduleEvidenceScenariosRequiresLaunchdLiveState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schedule.json")
	body := `{"status":"passed","installed":true,"s3_prefix_configured":true,"coverage_evidence":"coverage.json"}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := scheduleEvidenceScenarios(path)
	if len(got) != 3 || got[0].Status != statusFailed {
		t.Fatalf("scenarios=%+v", got)
	}
}

func TestScheduleEvidenceScenariosRequiresSuccessfulLaunchdRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schedule.json")
	body := `{"status":"passed","installed":true,"s3_prefix_configured":true,"coverage_evidence":"coverage.json","launchd":{"checked":true,"loaded":true,"calendar_trigger":true,"runs":0,"last_exit_code":"(never exited)"}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := scheduleEvidenceScenarios(path)
	if len(got) != 3 || got[0].Status != statusFailed {
		t.Fatalf("scenarios=%+v", got)
	}
}

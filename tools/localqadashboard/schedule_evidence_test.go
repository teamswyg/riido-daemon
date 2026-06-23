package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScheduleEvidenceScenariosRejectsTokenText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schedule.json")
	body := `{"status":"passed","installed":true,"s3_prefix_configured":true,"coverage_evidence":"coverage.json","command_has_token_text":true,"launchd":{"checked":true,"loaded":true,"calendar_trigger":true}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := scheduleEvidenceScenarios(path)
	if len(got) != 1 || got[0].Status != "failed" {
		t.Fatalf("scenarios=%+v", got)
	}
}

func TestScheduleEvidenceScenariosIncludesInstallDetail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schedule.json")
	body := `{"status":"passed","label":"io.riido.local-qa","installed":true,"hour":9,"minute":0,"s3_prefix_configured":true,"coverage_evidence":"coverage.json","command_preview":"go run ./tools/localqarunner","launchd":{"checked":true,"loaded":true,"state":"not running","runs":1,"last_exit_code":"0","calendar_trigger":true}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := scheduleEvidenceScenarios(path)
	if len(got) != 1 || got[0].FailureSummary == "" {
		t.Fatalf("scenarios=%+v", got)
	}
	if got[0].Status != statusPassed {
		t.Fatalf("status=%s", got[0].Status)
	}
	if got[0].Evidence != path {
		t.Fatalf("schedule provenance missing: %+v", got[0])
	}
	if !strings.Contains(got[0].FailureSummary, "launchd_loaded=true") {
		t.Fatalf("summary=%q", got[0].FailureSummary)
	}
	if !strings.Contains(got[0].FailureSummary, "calendar_trigger=true") {
		t.Fatalf("summary=%q", got[0].FailureSummary)
	}
	if !strings.Contains(got[0].FailureSummary, "coverage=coverage.json") {
		t.Fatalf("summary=%q", got[0].FailureSummary)
	}
}

func TestScheduleEvidenceScenariosRequiresLaunchdLiveState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schedule.json")
	body := `{"status":"passed","installed":true,"s3_prefix_configured":true,"coverage_evidence":"coverage.json"}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := scheduleEvidenceScenarios(path)
	if len(got) != 1 || got[0].Status != statusFailed {
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
	if len(got) != 1 || got[0].Status != statusFailed {
		t.Fatalf("scenarios=%+v", got)
	}
}

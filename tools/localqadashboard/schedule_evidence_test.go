package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScheduleEvidenceScenariosRejectsTokenText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schedule.json")
	body := `{"status":"passed","installed":true,"s3_prefix_configured":true,"coverage_evidence":"coverage.json","command_has_token_text":true,"trigger":{"id":"daily-evidence-sweep"},"launchd":{"checked":true,"loaded":true,"calendar_trigger":true}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := scheduleEvidenceScenarios(path)
	if len(got) != 3 || got[0].Status != "failed" {
		t.Fatalf("scenarios=%+v", got)
	}
}

func TestScheduleEvidenceScenariosIncludesInstallDetail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schedule.json")
	body := `{"status":"passed","label":"io.riido.local-qa","installed":true,"hour":9,"minute":0,"s3_prefix_configured":true,"coverage_evidence":"coverage.json","command_preview":"go run ./tools/localqarunner","trigger":{"id":"daily-evidence-sweep","name":"Daily Evidence Sweep","common_name":"하루 한 번 증적 순회","cadence":"daily","time_local":"09:00","freshness_window":"24h","closed_loop":"write evidence with expires_at","refreshes_expired_evidence":true,"evidence":["coverage.json"]},"launchd":{"checked":true,"loaded":true,"state":"not running","runs":1,"last_exit_code":"0","calendar_trigger":true}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := scheduleEvidenceScenarios(path)
	if len(got) != 3 || got[0].FailureSummary == "" || got[1].FailureSummary == "" {
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
	if got[1].ID != "local.qa.daily_trigger" || got[1].Status != statusPassed {
		t.Fatalf("trigger scenario=%+v", got[1])
	}
	if !strings.Contains(got[1].FailureSummary, "common_name=하루 한 번 증적 순회") {
		t.Fatalf("summary=%q", got[1].FailureSummary)
	}
	if got[2].ID != "local.qa.daily_freshness" || got[2].Status != statusPassed {
		t.Fatalf("freshness scenario=%+v", got[2])
	}
}

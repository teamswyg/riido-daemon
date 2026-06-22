package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScheduleEvidenceScenariosRejectsTokenText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schedule.json")
	body := `{"status":"passed","s3_prefix_configured":true,"command_has_token_text":true}`
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
	body := `{"status":"passed","label":"io.riido.local-qa","hour":9,"minute":0,"s3_prefix_configured":true,"command_preview":"go run ./tools/localqarunner"}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := scheduleEvidenceScenarios(path)
	if len(got) != 1 || got[0].FailureSummary == "" {
		t.Fatalf("scenarios=%+v", got)
	}
}

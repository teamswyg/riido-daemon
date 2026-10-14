package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScheduleEvidenceScenariosRequiresFreshnessRefreshClaim(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schedule.json")
	body := `{"status":"passed","installed":true,"s3_prefix_configured":true,"coverage_evidence":"coverage.json","trigger":{"id":"daily-evidence-sweep","freshness_window":"24h","closed_loop":"write evidence without timestamp","evidence":["coverage.json"]},"launchd":{"checked":true,"loaded":true,"state":"not running","runs":1,"last_exit_code":"0","calendar_trigger":true}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := scheduleEvidenceScenarios(path)
	if len(got) != 3 || got[2].Status != statusFailed {
		t.Fatalf("scenarios=%+v", got)
	}
	if !strings.Contains(got[2].FailureSummary, "refreshes_expired_evidence=false") {
		t.Fatalf("summary=%q", got[2].FailureSummary)
	}
}

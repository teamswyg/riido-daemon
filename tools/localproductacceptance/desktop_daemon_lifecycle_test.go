package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDesktopDaemonLifecyclePromotesAppQuitWithoutSocket(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := desktopDaemonStopEvidencePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"observed_at":"2026-06-24T06:42:17Z","reason":"app-quit","method":"socket","profile":"production"}`
	if err := os.WriteFile(path, []byte(body+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := desktopDaemonLifecycleScenario()
	if got.Status != statusPartial || got.Repair == nil {
		t.Fatalf("scenario did not promote shutdown candidate: %+v", got)
	}
	if got.Repair.Class != "daemon_lifecycle_policy" {
		t.Fatalf("repair class = %q", got.Repair.Class)
	}

	gap := evidenceGapScenario([]scenario{got}, testEvidenceGapConfig(t))
	if !hasEvidenceGapCandidate(
		gap.Observed["closed_loop_candidates"].([]evidenceGapCandidate),
		"repair-local.daemon.desktop_shutdown_lifecycle",
	) {
		t.Fatalf("shutdown candidate missing: %+v", gap.Observed)
	}
}

func TestDesktopDaemonLifecyclePassesWhenNoStopEvidence(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	got := desktopDaemonLifecycleScenario()
	if got.Status != statusPassed {
		t.Fatalf("scenario = %+v", got)
	}
}

func testEvidenceGapConfig(t *testing.T) config {
	t.Helper()
	manualOut := filepath.Join(t.TempDir(), "manual.json")
	screenshots := filepath.Join(t.TempDir(), "screenshots")
	return config{manualOut: &manualOut, screenshots: &screenshots}
}

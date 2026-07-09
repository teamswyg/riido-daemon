package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteScheduleEvidenceRedactsTokenCommandAndUsesRelativePath(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()
	relative := "nested/schedule.json"
	cfg.evidenceOut = &relative
	install := false
	cfg.install = &install
	paths := schedulePaths{
		repo:   dir,
		plist:  filepath.Join(dir, "qa.plist"),
		stdout: filepath.Join(dir, "out.log"),
		stderr: filepath.Join(dir, "err.log"),
	}
	live := launchdEvidence{Loaded: true, CalendarTrigger: true}

	err := writeScheduleEvidence(cfg, paths, "RIIDO_TOKEN=secret go run ./tools/localqarunner", live)
	if err != nil {
		t.Fatal(err)
	}

	var got scheduleEvidence
	body, err := os.ReadFile(filepath.Join(dir, relative))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if !got.Installed || !got.CommandHasTokenText {
		t.Fatalf("expected installed token evidence: %+v", got)
	}
	if got.CommandPreview != "[redacted: command contains token text]" {
		t.Fatalf("command preview leaked: %q", got.CommandPreview)
	}
	if got.Launchd.CalendarTrigger != true {
		t.Fatalf("launchd evidence missing: %+v", got.Launchd)
	}
	if _, err := os.Stat(filepath.Join(dir, "nested")); err != nil {
		t.Fatalf("relative evidence dir missing: %v", err)
	}
}

func TestRenderPlistCanRunAtLoad(t *testing.T) {
	cfg := testConfig()
	runAtLoad := true
	cfg.runAtLoad = &runAtLoad
	got := renderPlist(cfg, schedulePaths{repo: "/tmp/repo"})
	if !containsAll(got, "<key>RunAtLoad</key><true/>", "<key>StartCalendarInterval</key>") {
		t.Fatalf("plist missing run-at-load schedule:\n%s", got)
	}
}

func containsAll(text string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(text, part) {
			return false
		}
	}
	return true
}

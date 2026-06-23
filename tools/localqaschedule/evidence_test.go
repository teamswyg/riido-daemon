package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCommandMentionsToken(t *testing.T) {
	if !commandMentionsToken("RIIDO_AI_AGENT_TOKEN=x") {
		t.Fatal("token text not detected")
	}
	if commandMentionsToken("-product-storage-state .riido-local/private/state.json") {
		t.Fatal("storage-state should not be treated as token text")
	}
}

func TestSafeCommandPreviewRedactsTokenText(t *testing.T) {
	got := safeCommandPreview("RIIDO_AI_AGENT_TOKEN=x go run ./tools/localqarunner")
	if got != "[redacted: command contains token text]" {
		t.Fatalf("preview=%q", got)
	}
}

func TestWriteScheduleEvidenceIncludesCoveragePath(t *testing.T) {
	cfg := testConfig()
	out := filepath.Join(t.TempDir(), "schedule.json")
	cfg.evidenceOut = &out
	paths := schedulePaths{repo: t.TempDir(), plist: "/tmp/qa.plist"}
	err := writeScheduleEvidence(cfg, paths, localQACommand(cfg, paths), launchdEvidence{})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var got scheduleEvidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.CoverageEvidence != "/tmp/coverage.json" {
		t.Fatalf("coverage evidence=%q", got.CoverageEvidence)
	}
	if got.CommandHasTokenText {
		t.Fatal("command should not contain token text")
	}
}

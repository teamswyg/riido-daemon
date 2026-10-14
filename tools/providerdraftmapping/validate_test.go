package main

import (
	"path/filepath"
	"testing"
)

func TestValidateRejectsMappingDrift(t *testing.T) {
	repo := t.TempDir()
	mustWrite(t, filepath.Join(repo, "provider_event_draft.go"), sourceFixture())
	manifest := validManifest("doc.md")
	manifest.MappedEvents[0].EventTypeConst = "EventLogLine"
	problems, _, _ := validate(repo, manifest)
	if len(problems) == 0 {
		t.Fatalf("expected mapping drift failure")
	}
}

func TestValidateRejectsCoverageGap(t *testing.T) {
	repo := t.TempDir()
	mustWrite(t, filepath.Join(repo, "provider_event_draft.go"), sourceFixture())
	manifest := validManifest("doc.md")
	manifest.SkippedEvents = nil
	problems, _, _ := validate(repo, manifest)
	if len(problems) == 0 {
		t.Fatalf("expected coverage failure")
	}
}

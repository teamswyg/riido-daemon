package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestMentionsMappedSkippedAndMissingEvents(t *testing.T) {
	manifest := Manifest{
		MappedEvents:  []MappedEvent{{EventKind: "text_delta"}},
		SkippedEvents: []SkippedEvent{{EventKind: "unknown"}},
	}
	for _, eventKind := range []string{"text_delta", "unknown"} {
		if !manifestMentions(manifest, eventKind) {
			t.Fatalf("expected manifest to mention %s", eventKind)
		}
	}
	if manifestMentions(manifest, "extra") {
		t.Fatal("unexpected manifest mention for extra event")
	}
}

func TestValidateMappingsReportsSourceRowsMissingFromManifest(t *testing.T) {
	manifest := validManifest("doc.md")
	_, problems := validateMappings(manifest, map[string]string{
		"text_delta": "EventTextDelta",
		"extra":      "EventTextDelta",
	})
	if len(problems) == 0 || !strings.Contains(problemMessages(problems)[0], "source mapping missing from manifest: extra") {
		t.Fatalf("expected source mapping gap, got %#v", problems)
	}
}

func TestWriteJSONReportsMarshalAndWriteErrors(t *testing.T) {
	repo := t.TempDir()
	err := writeJSON(filepath.Join(repo, "bad.json"), map[string]any{"bad": func() {}})
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("expected marshal error, got %v", err)
	}
	err = writeJSON(repo, Manifest{ID: "x"})
	if err == nil || !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("expected write error, got %v", err)
	}
}

func TestLoadManifestRejectsUnsafeAndReadsValidManifest(t *testing.T) {
	repo := t.TempDir()
	if _, err := loadManifest(repo, "../manifest.json"); err == nil {
		t.Fatal("expected unsafe manifest path to fail")
	}
	dir := filepath.Join(repo, "nested")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWriteManifest(t, repo, filepath.Join("nested", "manifest.json"), validManifest("doc.md"))
	got, err := loadManifest(repo, filepath.Join("nested", "manifest.json"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if got.ID != "test" || got.Source != "provider_event_draft.go" {
		t.Fatalf("manifest=%+v", got)
	}
}

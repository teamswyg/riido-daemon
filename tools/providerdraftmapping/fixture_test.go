package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func validManifest(doc string) Manifest {
	return Manifest{
		SchemaVersion:    schemaVersion,
		ID:               "test",
		Title:            "Test",
		GeneratedDoc:     doc,
		Workflow:         "workflow.yml",
		EvidenceArtifact: "artifact",
		Source:           "provider_event_draft.go",
		MappedEvents:     []MappedEvent{{EventKind: "text_delta", EventKindConst: "EventTextDelta", EventTypeConst: "EventTextDelta", EventType: "TextDelta"}},
		SkippedEvents:    fixtureSkippedEvents("text_delta"),
		Assertions:       []string{"mapping matches source"},
	}
}

func mustWriteManifest(t *testing.T, repo, path string, manifest Manifest) {
	t.Helper()
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(repo, path), string(data))
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

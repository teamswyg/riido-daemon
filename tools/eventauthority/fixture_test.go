package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func validManifest(doc string) Manifest {
	return Manifest{
		SchemaVersion:          schemaVersion,
		ID:                     "test",
		Title:                  "Test",
		GeneratedDoc:           doc,
		Workflow:               "workflow.yml",
		EvidenceArtifact:       "artifact",
		DraftSource:            "internal/ir/ingest/draft.go",
		BuilderSource:          "internal/ir/ingest/event_from_draft.go",
		DraftSuppliedFields:    []string{"OccurredAt", "Scope", "Type", "Payload", "Unknown", "TaskID"},
		IngestorAssignedFields: []string{"EventID", "EventSchemaVersion", "ActorKind", "ActorID"},
		Rules:                  []string{"single append API"},
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

package project

import (
	"path/filepath"
	"testing"
)

func TestSaveAndLoadState(t *testing.T) {
	projection, err := FromMwsdSnapshot(sampleSnapshot())
	if err != nil {
		t.Fatalf("FromMwsdSnapshot returned error: %v", err)
	}
	path := filepath.Join(t.TempDir(), "workspace-state.json")

	if err := SaveState(path, StateFromProjection(projection)); err != nil {
		t.Fatalf("SaveState returned error: %v", err)
	}
	loaded, err := LoadState(path)
	if err != nil {
		t.Fatalf("LoadState returned error: %v", err)
	}
	if loaded.SchemaVersion != StateSchemaVersion {
		t.Fatalf("unexpected loaded schema: %s", loaded.SchemaVersion)
	}
	if loaded.Tasks[1].ID != "task:mws.roadmap" {
		t.Fatalf("unexpected loaded task: %#v", loaded.Tasks[1])
	}
}

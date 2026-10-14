package main

import "testing"

func TestManifestSeparatesImplementedAndReservedActions(t *testing.T) {
	manifest, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.ImplementedAction) != 3 {
		t.Fatalf("implemented actions = %d, want 3", len(manifest.ImplementedAction))
	}
	if len(manifest.ReservedActions) != 2 {
		t.Fatalf("reserved actions = %d, want 2", len(manifest.ReservedActions))
	}
}

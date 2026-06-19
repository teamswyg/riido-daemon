package main

import "testing"

func TestManifestHasFactEvidence(t *testing.T) {
	manifest, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Facts) < 6 {
		t.Fatalf("want multiple executable facts, got %d", len(manifest.Facts))
	}
	if len(manifest.AbsentSurfaces) == 0 {
		t.Fatal("expected absent surface checks")
	}
}

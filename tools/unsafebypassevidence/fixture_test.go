package main

import "testing"

func TestManifestCoversUnsafeBypassSurfaces(t *testing.T) {
	manifest, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Surfaces) != 4 {
		t.Fatalf("surfaces = %d, want 4", len(manifest.Surfaces))
	}
	if len(manifest.SourceChecks) < 12 {
		t.Fatalf("source checks = %d, want enforcement coverage", len(manifest.SourceChecks))
	}
}

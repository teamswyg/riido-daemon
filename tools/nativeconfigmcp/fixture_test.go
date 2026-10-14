package main

import "testing"

func TestManifestCoversHookFileAndMCPFacts(t *testing.T) {
	manifest, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Facts) < 6 {
		t.Fatalf("facts = %d, want hook, file, and MCP coverage", len(manifest.Facts))
	}
	if len(manifest.AbsentSurfaces) == 0 {
		t.Fatal("expected absent surfaces")
	}
}

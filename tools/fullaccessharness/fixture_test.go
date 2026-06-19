package main

import "testing"

func TestFullAccessHarnessFixtureHasAssertions(t *testing.T) {
	manifest, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Facts) < 6 {
		t.Fatalf("expected implemented facts, got %d", len(manifest.Facts))
	}
	if len(manifest.AbsentSurfaces) == 0 {
		t.Fatal("expected absent-surface checks")
	}
}

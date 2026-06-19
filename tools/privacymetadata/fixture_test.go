package main

import "testing"

func TestPrivacyMetadataFixtureUsesAllowlistArtifact(t *testing.T) {
	manifest, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.PolicyArtifact != "internal/hostintegration/privacy_metadata_allowlist.riido.json" {
		t.Fatalf("policy artifact = %q", manifest.PolicyArtifact)
	}
}

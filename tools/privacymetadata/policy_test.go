package main

import "testing"

func TestPrivacyMetadataPolicySnapshot(t *testing.T) {
	manifest, err := loadManifest("../..", defaultManifest)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := loadPolicy("../..", manifest.PolicyArtifact)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := findSurface(policy, serverFacingSurfaceID); !ok {
		t.Fatal("missing server-facing metadata surface")
	}
	if _, ok := findSurface(policy, providerStatusSurfaceID); !ok {
		t.Fatal("missing provider status sync surface")
	}
}

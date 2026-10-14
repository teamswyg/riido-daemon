package main

import "testing"

func TestBuildEvidenceIncludesManifestInventory(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "docs/a.riido.json", "{}\n")
	writeFixture(t, root, "packaging/b.riido.json", "{}\n")
	evidence := buildEvidence(root, fixtureManifest(), nil, nil)
	if evidence.ManifestInventory.Count != 2 {
		t.Fatalf("inventory count = %d", evidence.ManifestInventory.Count)
	}
	if len(evidence.ManifestInventory.Groups) != 2 || len(evidence.ManifestInventory.Samples) != 2 {
		t.Fatalf("inventory = %+v", evidence.ManifestInventory)
	}
}

package main

import "testing"

func TestPrivacyMetadataEvidenceClean(t *testing.T) {
	if err := run(options{Repo: "../..", Manifest: defaultManifest, CheckDoc: true}); err != nil {
		t.Fatal(err)
	}
}

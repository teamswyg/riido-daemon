package main

import (
	"path/filepath"
	"testing"
)

func TestRunChecksCurrentManifest(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	err := run(options{Repo: "../..", Manifest: defaultManifest, CheckDoc: true, EvidenceOut: out})
	if err != nil {
		t.Fatal(err)
	}
}

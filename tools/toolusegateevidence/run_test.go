package main

import (
	"path/filepath"
	"testing"
)

func TestRunChecksCurrentManifest(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := run(options{Repo: "../..", Manifest: defaultManifest, CheckDoc: true, EvidenceOut: out}); err != nil {
		t.Fatal(err)
	}
}

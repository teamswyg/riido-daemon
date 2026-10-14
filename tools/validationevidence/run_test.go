package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentManifest(t *testing.T) {
	if err := run(options{Repo: "../..", Manifest: "docs/20-domain/validation.riido.json"}); err != nil {
		t.Fatal(err)
	}
}

func TestGeneratedDocCurrent(t *testing.T) {
	err := run(options{Repo: "../..", Manifest: "docs/20-domain/validation.riido.json", CheckDoc: true})
	if err != nil {
		t.Fatal(err)
	}
}

func TestFixtureCanWriteAndCheck(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir)
	opts := options{Repo: dir, Manifest: "manifest.json", WriteDoc: true, CheckDoc: true}
	if err := run(opts); err != nil {
		t.Fatal(err)
	}
}

func TestRejectsSourceDrift(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir)
	mustWrite(t, filepath.Join(dir, "src.go"), "other")
	err := run(options{Repo: dir, Manifest: "manifest.json"})
	if err == nil || !strings.Contains(err.Error(), "source check failed") {
		t.Fatalf("expected source check error, got %v", err)
	}
}

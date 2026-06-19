package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentManifestAndGeneratedDoc(t *testing.T) {
	if err := run("../..", "docs/readme/verification.riido.json", "", false, true, false); err != nil {
		t.Fatal(err)
	}
}

func TestRejectsDuplicateCommandID(t *testing.T) {
	dir, manifestPath := fixture(t)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	mutated := strings.Replace(string(data), `"id":"b"`, `"id":"a"`, 1)
	mustWrite(t, manifestPath, mutated)
	err = run(dir, manifestPath, "", false, false, false)
	if err == nil || !strings.Contains(err.Error(), `duplicate command id "a"`) {
		t.Fatalf("expected duplicate id error, got %v", err)
	}
}

func TestEvidenceWithoutCommandRunIsVerified(t *testing.T) {
	dir, manifestPath := fixture(t)
	evidencePath := filepath.Join(dir, "evidence.json")
	if err := run(dir, manifestPath, evidencePath, true, true, false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"status": "verified"`) {
		t.Fatalf("expected verified evidence, got %s", data)
	}
}

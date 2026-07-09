package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestAndDocChecks(t *testing.T) {
	root := t.TempDir()
	bad := filepath.Join(root, "bad.json")
	mustWrite(t, bad, "{")
	if _, err := loadManifest(bad); err == nil || !strings.Contains(err.Error(), "decode manifest") {
		t.Fatalf("expected decode error, got %v", err)
	}
	manifestPath := writeFixtureManifest(t, root)
	m, err := loadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	docPath := filepath.Join(root, "self.md")
	if err := checkDoc(docPath, m); err == nil || !strings.Contains(err.Error(), "read generated doc") {
		t.Fatalf("expected missing doc error, got %v", err)
	}
	if err := writeDoc(docPath, m); err != nil {
		t.Fatal(err)
	}
	if err := checkDoc(docPath, m); err != nil {
		t.Fatalf("fresh doc drifted: %v", err)
	}
	mustWrite(t, docPath, "stale")
	if err := checkDoc(docPath, m); err == nil || !strings.Contains(err.Error(), "generated doc drift") {
		t.Fatalf("expected doc drift, got %v", err)
	}
}

func TestValidateManifestClosedLoopFailures(t *testing.T) {
	root := t.TempDir()
	manifestPath := writeFixtureManifest(t, root)
	body := string(mustRead(t, manifestPath))
	body = strings.Replace(body, `"id":"feature"`, `"id":"bug"`, 1)
	mustWrite(t, manifestPath, body)
	if _, err := loadManifest(manifestPath); err == nil || !strings.Contains(err.Error(), "duplicate closed loop bug") {
		t.Fatalf("expected duplicate closed loop error, got %v", err)
	}
}

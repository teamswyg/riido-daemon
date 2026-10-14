package main

import (
	"strings"
	"testing"
)

func TestManifestRejectsProducerCommandWithoutExpectedOutput(t *testing.T) {
	root := t.TempDir()
	manifestPath := writeFixtureManifest(t, root)
	body := string(mustRead(t, manifestPath))
	body = strings.Replace(body, "<evidence-dir>/loop.json", "wrong.json", 1)
	mustWrite(t, manifestPath, body)
	_, err := loadManifest(manifestPath)
	if err == nil {
		t.Fatal("expected invalid producer command")
	}
	if !strings.Contains(err.Error(), "producer_command must write") {
		t.Fatalf("unexpected error: %v", err)
	}
}

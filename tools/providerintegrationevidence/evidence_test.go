package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestDocCheckDoesNotProbeProviderExecutables(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell probe fixture is unix-only")
	}
	dir, manifestPath, docPath := newFixture(t)
	marker := filepath.Join(dir, "probed")
	probe := filepath.Join(dir, "probe")
	mustWrite(t, probe, "#!/bin/sh\ntouch '"+marker+"'\n")
	if err := os.Chmod(probe, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RIIDO_FAKE_PROVIDER_PATH", probe)
	mustWrite(t, docPath, renderMarkdown(mustLoad(t, manifestPath)))
	if err := run(dir, manifestPath, "", false, true, false, 24*time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("doc check probed provider executable, stat err=%v", err)
	}
}

func TestEvidenceOnlyRecordsObservedStatus(t *testing.T) {
	dir, manifestPath, docPath := newFixture(t)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("RIIDO_FAKE_PROVIDER_PATH", exe)
	mustWrite(t, docPath, renderMarkdown(mustLoad(t, manifestPath)))
	evidencePath := filepath.Join(dir, "evidence.json")
	if err := run(dir, manifestPath, evidencePath, false, true, false, 24*time.Hour); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatal(err)
	}
	var evidence evidenceFile
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.Status != "observed" {
		t.Fatalf("status=%q, want observed", evidence.Status)
	}
	if evidence.Providers[0].ExecutableRef != "$RIIDO_FAKE_PROVIDER_PATH" {
		t.Fatalf("executable_ref=%q", evidence.Providers[0].ExecutableRef)
	}
	if evidence.Providers[0].ExecutablePath != exe {
		t.Fatalf("executable_path=%q, want %q", evidence.Providers[0].ExecutablePath, exe)
	}
	if evidence.Providers[0].IntegrationStatus != "observed" {
		t.Fatalf("integration_status=%q", evidence.Providers[0].IntegrationStatus)
	}
	assertFreshEvidence(t, evidence)
}

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestObserveProviderMissingExecutableCarriesRepair(t *testing.T) {
	ev := observeProvider(t.TempDir(), provider{
		ID:                "missing",
		DefaultExecutable: "definitely-missing-riido-provider",
		OverrideEnv:       "RIIDO_MISSING_PROVIDER_PATH",
		GoPackage:         ".",
		TestRegex:         "TestIntegration",
	}, true)
	if ev.Available || ev.IntegrationStatus != "skipped" {
		t.Fatalf("evidence=%+v", ev)
	}
	if ev.Repair == nil || ev.Repair.Class != "provider_executable_missing" {
		t.Fatalf("repair=%+v", ev.Repair)
	}
}

func TestObserveProviderWithoutIntegrationRecordsObserved(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("RIIDO_FAKE_PROVIDER_PATH", exe)
	ev := observeProvider(t.TempDir(), provider{
		ID:                "fake",
		DefaultExecutable: "missing-fake",
		OverrideEnv:       "RIIDO_FAKE_PROVIDER_PATH",
		GoPackage:         ".",
		TestRegex:         "TestIntegration",
	}, false)
	if !ev.Available || ev.IntegrationStatus != "observed" {
		t.Fatalf("evidence=%+v", ev)
	}
	if ev.ExecutableRef != "$RIIDO_FAKE_PROVIDER_PATH" || ev.Repair != nil {
		t.Fatalf("ref=%q repair=%+v", ev.ExecutableRef, ev.Repair)
	}
}

func TestRunReportsManifestAndEvidenceWriteErrors(t *testing.T) {
	dir, manifestPath, docPath := newFixture(t)
	mustWrite(t, docPath, renderMarkdown(mustLoad(t, manifestPath)))
	if err := run(dir, "{bad-path}", "", false, false, false, time.Hour); err == nil ||
		!strings.Contains(err.Error(), "read manifest") {
		t.Fatalf("expected manifest read error, got %v", err)
	}
	err := run(dir, manifestPath, dir, false, true, false, time.Hour)
	if err == nil || !strings.Contains(err.Error(), "write evidence") {
		t.Fatalf("expected evidence write error, got %v", err)
	}
}

func TestWriteJSONReportsEncodeAndWriteErrors(t *testing.T) {
	dir := t.TempDir()
	err := writeJSON(filepath.Join(dir, "bad.json"), map[string]any{"bad": func() {}})
	if err == nil || !strings.Contains(err.Error(), "encode evidence") {
		t.Fatalf("expected encode error, got %v", err)
	}
	err = writeJSON(dir, map[string]string{"status": "ok"})
	if err == nil || !strings.Contains(err.Error(), "write evidence") {
		t.Fatalf("expected write error, got %v", err)
	}
}

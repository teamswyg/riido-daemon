package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyInstallerFailsWhenOldBinarySurvivesCopyBoundary(t *testing.T) {
	fixture, cleanup, err := newInstallFixture()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	repo := t.TempDir()
	writeInstallScript(t, repo, `#!/bin/sh
set -eu
archive="$INSTALL_FIXTURE_DIR/riido-daemon_darwin_arm64.tar.gz"
install -m 0755 "$archive" "$RIIDO_DAEMON_INSTALL_DIR/riido"
`)
	scenario, _ := verifyInstaller(t.Context(), repo, fixture)
	if scenario.Status != statusFailed || !strings.Contains(scenario.FailureSummary, "old daemon") {
		t.Fatalf("scenario = %+v", scenario)
	}
}

func TestAggregateStatusFailsWhenAnyScenarioFails(t *testing.T) {
	got := aggregateStatus([]scenario{{Status: statusPassed}, {Status: statusFailed}})
	if got != statusFailed {
		t.Fatalf("aggregate status = %q", got)
	}
	if got := aggregateStatus([]scenario{{Status: statusPassed}}); got != statusPassed {
		t.Fatalf("aggregate passed status = %q", got)
	}
}

func TestOutputPathAndWriteJSONFailures(t *testing.T) {
	dir := t.TempDir()
	abs := filepath.Join(dir, "absolute.json")
	if got := outputPath("/ignored", abs); got != abs {
		t.Fatalf("absolute output path = %q", got)
	}
	err := writeJSON(filepath.Join(dir, "bad.json"), map[string]any{"fn": func() {}})
	if err == nil || !strings.Contains(err.Error(), "marshal evidence") {
		t.Fatalf("expected marshal error, got %v", err)
	}
}

func writeInstallScript(t *testing.T, repo, body string) {
	t.Helper()
	path := filepath.Join(repo, "scripts", "install-riido-daemon.sh")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

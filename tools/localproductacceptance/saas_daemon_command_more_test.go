package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaaSDaemonCommandScenariosUseConfiguredBinary(t *testing.T) {
	binary, logPath := fakeSaaSDaemonBinary(t)
	runDir := t.TempDir()
	cfg := config{daemonRunDir: &runDir}
	slot := newSaaSDaemonSlot(cfg, 1, saasDeviceCredential{
		DeviceID: "dev-1", DeviceSecret: "secret", DisplayName: "QA",
	})

	start := startSaaSDaemon(binary, slot, "https://staging.ai-api.riido.io")
	if start.Status != statusPassed || start.Observed["device_id"] != "dev-1" {
		t.Fatalf("start=%+v", start)
	}
	status := readSaaSDaemonStatus(binary, slot)
	if status.Status != statusPassed || status.Observed["ready"] != true ||
		status.Observed["runtimes_count"] != 1 {
		t.Fatalf("status=%+v", status)
	}
	cleanup := cleanupSaaSDaemons(binary, saasPrepareResult{Slots: []saasDaemonSlot{slot}})
	if len(cleanup) != 1 || cleanup[0].Status != statusPassed {
		t.Fatalf("cleanup=%+v", cleanup)
	}
	stopExistingSaaSDaemon(binary, slot)

	logged, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"daemon start --socket " + slot.Socket,
		"daemon status --socket " + slot.Socket,
		"daemon stop --socket " + slot.Socket,
	} {
		if !strings.Contains(string(logged), want) {
			t.Fatalf("missing %q in:\n%s", want, logged)
		}
	}
}

func fakeSaaSDaemonBinary(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	binary := filepath.Join(dir, "riido")
	logPath := filepath.Join(dir, "calls.log")
	script := fmt.Sprintf(`#!/bin/sh
echo "$@" >> %q
if [ "$1 $2" = "daemon status" ]; then
  echo '{"ready":true,"daemon_version":"v-test","profile":"staging","runtimes":[{"kind":"codex"}]}'
else
  echo ok
fi
`, logPath)
	if err := os.WriteFile(binary, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return binary, logPath
}

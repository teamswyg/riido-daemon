package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDaemonBackgroundStartLaunchesChildAndReturns(t *testing.T) {
	paths := newDaemonBackgroundPaths(t)
	started := time.Now()
	if err := run(paths.startArgs()); err != nil {
		t.Fatalf("background start: %v", err)
	}
	if elapsed := time.Since(started); elapsed > 8*time.Second {
		t.Fatalf("background start blocked too long: %v", elapsed)
	}
	cleanupBackgroundDaemon(t, paths)
	assertBackgroundDaemonHealthy(t, paths.socket)
	assertBackgroundPIDIdentity(t, paths, readDaemonPIDFile(t, paths.pid))
}

func assertBackgroundDaemonHealthy(t *testing.T, socket string) {
	t.Helper()
	out, err := runCapturingStdout(t, func() error {
		return run([]string{"daemon", "status", "--socket", socket})
	})
	if err != nil {
		t.Fatalf("status: %v\n%s", err, out)
	}
	var status struct {
		Health string `json:"health"`
	}
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		t.Fatalf("parse status: %v\n%s", err, out)
	}
	if status.Health != "ok" {
		t.Fatalf("health: %q\nfull: %s", status.Health, out)
	}
}

package main

import "testing"

func assertDaemonStatusFields(t *testing.T, status daemonStatusJSON, sock string) {
	t.Helper()
	if status.SchemaVersion == "" {
		t.Fatalf("schema_version missing")
	}
	if status.DaemonID != "daemon-test-1" {
		t.Fatalf("daemon_id mismatch: %q", status.DaemonID)
	}
	if status.DaemonVersion != "riido-agentd v1.2.3" {
		t.Fatalf("daemon_version mismatch: %q", status.DaemonVersion)
	}
	if status.Profile != "desktop-api.riido.ai" || status.ServerURL != "https://api.riido.ai" {
		t.Fatalf("profile/server_url mismatch: profile=%q server=%q", status.Profile, status.ServerURL)
	}
	if status.DeviceName != "MacBook-Pro-SK.local" || status.WorkspaceCount != 2 {
		t.Fatalf("device/workspace mismatch: device=%q workspace=%d", status.DeviceName, status.WorkspaceCount)
	}
	if status.PID == 0 {
		t.Fatalf("pid missing")
	}
	if status.Health != "ok" {
		t.Fatalf("health: %q", status.Health)
	}
	if !status.Ready || status.Readiness != "ready" {
		t.Fatalf("ready: ready=%v readiness=%q", status.Ready, status.Readiness)
	}
	if status.SocketPath != sock {
		t.Fatalf("socket path mismatch: %q vs %q", status.SocketPath, sock)
	}
}

func assertDaemonMetrics(t *testing.T, status daemonStatusJSON) {
	t.Helper()
	if status.Metrics.RuntimeCount != 4 || status.Metrics.RuntimeResponding != 4 {
		t.Fatalf("runtime metrics mismatch: %+v", status.Metrics)
	}
	if status.Metrics.ProviderAvailable+status.Metrics.ProviderUnavailable != 4 {
		t.Fatalf("provider metrics mismatch: %+v", status.Metrics)
	}
	if status.Metrics.RunningTasks != status.RunningTasks {
		t.Fatalf("running task metrics mismatch: metrics=%d status=%d", status.Metrics.RunningTasks, status.RunningTasks)
	}
}

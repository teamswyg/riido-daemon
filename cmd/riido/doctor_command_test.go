package main

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func doctorChecks(t *testing.T, env contextcli.Envelope) map[string]string {
	t.Helper()
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("doctor data: %#v", env.Data)
	}
	items, _ := data["checks"].([]any)
	out := map[string]string{}
	for _, item := range items {
		check, _ := item.(map[string]any)
		name, _ := check["name"].(string)
		status, _ := check["status"].(string)
		out[name] = status
	}
	return out
}

func TestDoctorReportsOldDaemonWithoutContextRead(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.daemon.replies[string(daemonMethodStatus)] = `{"schema_version":"1","daemon_version":"riido-agentd test"}`
	if got := exitCodeOf(t, runDoctor(h.io, []string{"--json"})); got != 8 {
		t.Fatalf("agent-mode broker is unsupported: want exit 8, got %d", got)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	data := env.Data.(map[string]any)
	if data["healthy"] != false || data["api_verified"] != false {
		t.Fatalf("unsupported agent broker cannot be healthy: %+v", data)
	}
	checks := doctorChecks(t, env)
	want := map[string]string{
		"binary": "ok", "path": "warn", "mode": "ok", "daemon_ipc": "ok",
		"context_read": "unsupported", "user_auth": "skip", "api": "skip",
	}
	for name, status := range want {
		if checks[name] != status {
			t.Fatalf("%s: want %s, got %s (all=%v)", name, status, checks[name], checks)
		}
	}
}

func TestDoctorUserModeSkipsAgentBrokerWithoutFailingLocalDiagnosis(t *testing.T) {
	h := newContextHarness(t, nil)
	h.daemon.replies[string(daemonMethodStatus)] = `{"daemon_version":"test"}`
	if err := runDoctor(h.io, []string{"--json"}); err != nil {
		t.Fatalf("user mode does not need an agent capability: %v", err)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	if data := env.Data.(map[string]any); data["healthy"] != true || data["api_verified"] != false {
		t.Fatalf("local-only user diagnosis: %+v", data)
	}
	if got := doctorChecks(t, env)["context_read"]; got != "skip" {
		t.Fatalf("user broker check: want skip, got %s", got)
	}
	if len(h.daemon.requests) != 1 {
		t.Fatal("user doctor must only request daemon status")
	}
}

func TestDoctorFailsWhenAgentCannotReachDaemon(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.daemon.down = true
	if got := exitCodeOf(t, runDoctor(h.io, []string{"--json"})); got != 6 {
		t.Fatalf("want exit 6, got %d", got)
	}
	checks := doctorChecks(t, decodeEnvelope(t, h.stdout.Bytes()))
	if checks["daemon_ipc"] != "fail" || checks["context_read"] != "skip" {
		t.Fatalf("unexpected checks: %v", checks)
	}
}

func TestDoctorOnlyWarnsInUserModeWithoutDaemon(t *testing.T) {
	h := newContextHarness(t, nil)
	h.daemon.down = true
	if err := runDoctor(h.io, []string{"--json"}); err != nil {
		t.Fatalf("user-mode doctor should stay healthy without daemon: %v", err)
	}
	if checks := doctorChecks(t, decodeEnvelope(t, h.stdout.Bytes())); checks["daemon_ipc"] != "warn" {
		t.Fatalf("unexpected checks: %v", checks)
	}
}

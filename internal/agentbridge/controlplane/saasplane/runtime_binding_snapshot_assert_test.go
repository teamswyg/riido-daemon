package saasplane

import (
	"testing"
	"time"
)

func assertRuntimeBindingSnapshot(t *testing.T, snapshot DeviceRuntimeSnapshotSyncRequest, startedAt time.Time) {
	t.Helper()
	if snapshot.DaemonID != "daemon-1" ||
		snapshot.DeviceID != "device-1" ||
		snapshot.DeviceDisplayName != "주윤의 MacBook" {
		t.Fatalf("heartbeat snapshot identity = %+v", snapshot)
	}
	if snapshot.Profile != "development" ||
		snapshot.AppVersion != "v0.0.13" ||
		snapshot.PID != 8765 ||
		!snapshot.StartedAt.Equal(startedAt) ||
		snapshot.UptimeSeconds <= 0 {
		t.Fatalf("heartbeat daemon facts = %+v", snapshot)
	}
	if len(snapshot.Runtimes) != 2 ||
		snapshot.Runtimes[0].RuntimeID != "daemon-1:codex" ||
		snapshot.Runtimes[1].RuntimeID != "daemon-1:cursor" {
		t.Fatalf("heartbeat snapshot must aggregate sorted runtimes: %+v", snapshot.Runtimes)
	}
	codex := snapshot.Runtimes[0]
	if len(codex.Models) != 1 ||
		codex.Models[0].ModelID != "gpt-5.5" ||
		codex.ProviderVersion != "codex-cli 0.133.0" ||
		!codex.RequiresExperimentalOptIn {
		t.Fatalf("codex runtime facts lost in heartbeat snapshot: %+v", codex)
	}
}

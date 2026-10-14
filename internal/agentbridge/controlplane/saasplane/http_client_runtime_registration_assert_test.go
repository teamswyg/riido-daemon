package saasplane

import (
	"testing"
	"time"
)

func assertRuntimeSnapshotIdentity(t *testing.T, snapshot DeviceRuntimeSnapshotSyncRequest, startedAt time.Time) {
	t.Helper()
	if snapshot.DaemonID != "daemon-1" ||
		snapshot.DeviceID != "device-1" ||
		snapshot.DeviceDisplayName != "주윤의 MacBook" {
		t.Fatalf("snapshot identity = %+v", snapshot)
	}
	if snapshot.Profile != "development" ||
		snapshot.AppVersion != "v0.0.13" ||
		snapshot.PID != 4321 ||
		!snapshot.StartedAt.Equal(startedAt) ||
		snapshot.UptimeSeconds <= 0 {
		t.Fatalf("snapshot daemon facts = %+v", snapshot)
	}
}

func assertRuntimeSnapshotRuntime(t *testing.T, runtimes []RuntimeSnapshotRecord) {
	t.Helper()
	if len(runtimes) != 1 ||
		runtimes[0].RuntimeID != "daemon-1:codex" ||
		runtimes[0].Kind != "codex" ||
		runtimes[0].Availability != "online" ||
		runtimes[0].DetectionState != "detected" ||
		runtimes[0].ProviderVersion != "codex-cli 0.133.0" ||
		!runtimes[0].RequiresExperimentalOptIn {
		t.Fatalf("snapshot runtimes = %+v", runtimes)
	}
	models := runtimes[0].Models
	if len(models) != 1 || models[0].ModelID != "gpt-5.5" || !models[0].IsDefault {
		t.Fatalf("snapshot runtime models = %+v", models)
	}
}

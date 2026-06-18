package taskdbplane

import (
	"testing"
	"time"
)

func assertRegisteredRuntime(t *testing.T, registry RuntimeRegistry, path string) {
	t.Helper()
	if registry.SchemaVersion != RuntimeRegistrySchemaVersion || registry.TaskDBPath != path {
		t.Fatalf("registry identity mismatch: %+v", registry)
	}
	if len(registry.Runtimes) != 1 || registry.Runtimes[0].RuntimeID != "runtime-1" {
		t.Fatalf("registry runtimes after register: %+v", registry.Runtimes)
	}
}

func assertRuntimeHeartbeat(t *testing.T, registry RuntimeRegistry, firstHeartbeat time.Time) {
	t.Helper()
	if len(registry.Runtimes) != 1 || !registry.Runtimes[0].LastHeartbeat.After(firstHeartbeat) {
		t.Fatalf("heartbeat was not persisted: before=%v registry=%+v", firstHeartbeat, registry.Runtimes)
	}
	if registry.Runtimes[0].SlotLimit != 2 || registry.Runtimes[0].SlotsInUse != 1 {
		t.Fatalf("slot heartbeat not persisted: %+v", registry.Runtimes[0])
	}
	if len(registry.Runtimes[0].RunningTaskIDs) != 2 || registry.Runtimes[0].RunningTaskIDs[0] != "task-a" {
		t.Fatalf("running task ids should be sorted: %+v", registry.Runtimes[0].RunningTaskIDs)
	}
}

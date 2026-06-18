package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestRuntimeRegistryPersistsRegistrationHeartbeatAndDeregister(t *testing.T) {
	path := writeTaskDB(t, taskdb.TaskDB{SchemaVersion: taskdb.TaskDBSchemaVersion})
	plane := newTestPlane(t, path)

	registerCodexRuntime(t, plane)
	registry := readRuntimeRegistry(t, plane.registryPath)
	assertRegisteredRuntime(t, registry, path)
	firstHeartbeat := registry.Runtimes[0].LastHeartbeat

	if err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      "runtime-1",
		SlotLimit:      2,
		SlotsInUse:     1,
		RunningTaskIDs: []string{"task-b", "task-a"},
	}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	registry = readRuntimeRegistry(t, plane.registryPath)
	assertRuntimeHeartbeat(t, registry, firstHeartbeat)

	if err := plane.DeregisterRuntime(context.Background(), "runtime-1"); err != nil {
		t.Fatalf("DeregisterRuntime: %v", err)
	}
	registry = readRuntimeRegistry(t, plane.registryPath)
	if len(registry.Runtimes) != 0 {
		t.Fatalf("registry runtimes after deregister: %+v", registry.Runtimes)
	}
}

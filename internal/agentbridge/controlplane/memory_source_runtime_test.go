package controlplane

import (
	"context"
	"testing"
	"time"
)

func newHeartbeatMemorySource() *MemorySource {
	src := NewMemorySource()
	now := time.Now()
	src.now = func() time.Time { return now }
	return src
}

func registerMemoryRuntime(t *testing.T, src *MemorySource) {
	t.Helper()
	reg := RuntimeRegistration{DaemonID: "d-1", RuntimeID: "rt-1", Provider: "claude"}
	if err := src.RegisterRuntime(context.Background(), reg); err != nil {
		t.Fatalf("Register: %v", err)
	}
}

func assertMemoryRuntimeRegistered(t *testing.T, src *MemorySource) {
	t.Helper()
	if rts := src.Registered(); len(rts) != 1 || rts[0].RuntimeID != "rt-1" {
		t.Fatalf("registered: %+v", rts)
	}
}

func heartbeatMemoryRuntime(t *testing.T, src *MemorySource) {
	t.Helper()
	now := src.now().Add(15 * time.Second)
	src.now = func() time.Time { return now }
	hb := RuntimeHeartbeat{RuntimeID: "rt-1", SlotLimit: 2, SlotsInUse: 1}
	hb.RunningTaskIDs = []string{"task-1"}
	if err := src.Heartbeat(context.Background(), hb); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	if got := src.Registered()[0].LastHeartbeat; !got.Equal(now) {
		t.Fatalf("heartbeat: %v", got)
	}
	if got := src.Registered()[0].SlotsInUse; got != 1 {
		t.Fatalf("slots in use after heartbeat = %d", got)
	}
}

func deregisterMemoryRuntime(t *testing.T, src *MemorySource) {
	t.Helper()
	if err := src.DeregisterRuntime(context.Background(), "rt-1"); err != nil {
		t.Fatalf("Deregister: %v", err)
	}
	if rts := src.Registered(); len(rts) != 0 {
		t.Fatalf("expected empty after deregister: %+v", rts)
	}
}

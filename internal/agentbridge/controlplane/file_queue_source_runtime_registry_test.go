package controlplane

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestFileQueueSourceWritesRuntimeRegistry(t *testing.T) {
	dir := t.TempDir()
	src, err := NewFileQueueSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	src.now = func() time.Time { return now }

	reg := RuntimeRegistration{
		DaemonID:     "daemon-1",
		RuntimeID:    "rt-1",
		Provider:     "multi",
		Capabilities: map[string]bool{"provider.codex.available": true},
		DeviceName:   "mac-mini",
		StartedAt:    now.Add(-time.Minute),
	}
	if err := src.RegisterRuntime(context.Background(), reg); err != nil {
		t.Fatalf("RegisterRuntime: %v", err)
	}

	rec := readFileRuntimeRecord(t, src.runtimePath("rt-1"))
	if rec.RuntimeID != "rt-1" || rec.DaemonID != "daemon-1" || rec.Provider != "multi" {
		t.Fatalf("runtime record identity: %+v", rec)
	}
	if !rec.LastHeartbeat.Equal(now) || !rec.Capabilities["provider.codex.available"] {
		t.Fatalf("runtime record state: %+v", rec)
	}

	now = now.Add(30 * time.Second)
	hb := RuntimeHeartbeat{RuntimeID: "rt-1", SlotLimit: 4, SlotsInUse: 2, RunningTaskIDs: []string{"task-b", "task-a"}}
	if err := src.Heartbeat(context.Background(), hb); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	rec = readFileRuntimeRecord(t, src.runtimePath("rt-1"))
	if rec.SlotLimit != 4 || rec.SlotsInUse != 2 {
		t.Fatalf("slot heartbeat mismatch: %+v", rec)
	}
	if len(rec.RunningTaskIDs) != 2 || rec.RunningTaskIDs[0] != "task-a" || rec.RunningTaskIDs[1] != "task-b" {
		t.Fatalf("running task ids should be sorted: %+v", rec.RunningTaskIDs)
	}

	if err := src.DeregisterRuntime(context.Background(), "rt-1"); err != nil {
		t.Fatalf("DeregisterRuntime: %v", err)
	}
	if _, err := os.Stat(src.runtimePath("rt-1")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("runtime record should be removed, stat err=%v", err)
	}
}

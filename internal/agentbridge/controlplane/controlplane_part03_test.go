package controlplane

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestFileQueueSourceSkipsUnavailableProviderForRuntime(t *testing.T) {
	dir := t.TempDir()
	src, err := NewFileQueueSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := src.RegisterRuntime(context.Background(), RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "rt-claude",
		Provider:  "multi",
		Capabilities: map[string]bool{
			"provider.claude.available": true,
			"provider.codex.available":  false,
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := src.RegisterRuntime(context.Background(), RuntimeRegistration{
		DaemonID:  "daemon-2",
		RuntimeID: "rt-codex",
		Provider:  "multi",
		Capabilities: map[string]bool{
			"provider.claude.available": false,
			"provider.codex.available":  true,
		},
	}); err != nil {
		t.Fatal(err)
	}
	for _, req := range []bridge.TaskRequest{
		{ID: "codex-task", Provider: "codex", Prompt: "x"},
		{ID: "claude-task", Provider: "claude", Prompt: "x"},
	} {
		body, _ := json.Marshal(req)
		if err := os.WriteFile(filepath.Join(dir, req.ID+".json"), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	first, err := src.ClaimTask(context.Background(), "rt-claude")
	if err != nil {
		t.Fatal(err)
	}
	if first == nil || first.ID != "claude-task" {
		t.Fatalf("rt-claude should skip codex and claim claude, got %+v", first)
	}
	if remaining := countTopLevelJSON(t, dir); remaining != 1 {
		t.Fatalf("codex task should remain for another runtime, remaining=%d", remaining)
	}

	second, err := src.ClaimTask(context.Background(), "rt-codex")
	if err != nil {
		t.Fatal(err)
	}
	if second == nil || second.ID != "codex-task" {
		t.Fatalf("rt-codex should claim remaining codex task, got %+v", second)
	}
	if claims := readClaimRecords(t, filepath.Join(dir, "claims")); len(claims) != 2 {
		t.Fatalf("claim receipts = %+v", claims)
	}
}

func TestFileQueueSourceClaimReceiptsDoNotOverwriteRepeatedFilenames(t *testing.T) {
	dir := t.TempDir()
	src, err := NewFileQueueSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 5, 25, 1, 2, 3, 0, time.UTC)
	src.now = func() time.Time { return now }

	writeTaskFile := func(id string) {
		t.Helper()
		req := bridge.TaskRequest{ID: id, Provider: "codex", Prompt: "x"}
		body, _ := json.Marshal(req)
		if err := os.WriteFile(filepath.Join(dir, "task.json"), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeTaskFile("first")
	if got, err := src.ClaimTask(context.Background(), "rt-1"); err != nil || got == nil || got.ID != "first" {
		t.Fatalf("first claim: %v / %+v", err, got)
	}
	writeTaskFile("second")
	if got, err := src.ClaimTask(context.Background(), "rt-1"); err != nil || got == nil || got.ID != "second" {
		t.Fatalf("second claim: %v / %+v", err, got)
	}

	claims := readClaimRecords(t, filepath.Join(dir, "claims"))
	if len(claims) != 2 {
		t.Fatalf("expected two claim receipts, got %+v", claims)
	}
	if claims[0].TaskID == claims[1].TaskID {
		t.Fatalf("claim receipts were overwritten: %+v", claims)
	}
}

func TestFileQueueSourceWritesRuntimeRegistry(t *testing.T) {
	dir := t.TempDir()
	src, err := NewFileQueueSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	src.now = func() time.Time { return now }

	reg := RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "rt-1",
		Provider:  "multi",
		Capabilities: map[string]bool{
			"provider.codex.available": true,
		},
		DeviceName: "mac-mini",
		StartedAt:  now.Add(-time.Minute),
	}
	if err := src.RegisterRuntime(context.Background(), reg); err != nil {
		t.Fatalf("RegisterRuntime: %v", err)
	}

	rec := readFileRuntimeRecord(t, src.runtimePath("rt-1"))
	if rec.RuntimeID != "rt-1" || rec.DaemonID != "daemon-1" || rec.Provider != "multi" {
		t.Fatalf("runtime record identity: %+v", rec)
	}
	if !rec.LastHeartbeat.Equal(now) {
		t.Fatalf("last heartbeat = %v, want %v", rec.LastHeartbeat, now)
	}
	if !rec.Capabilities["provider.codex.available"] {
		t.Fatalf("runtime capabilities missing: %+v", rec.Capabilities)
	}

	now = now.Add(30 * time.Second)
	if err := src.Heartbeat(context.Background(), RuntimeHeartbeat{RuntimeID: "rt-1", SlotLimit: 4, SlotsInUse: 2, RunningTaskIDs: []string{"task-b", "task-a"}}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	rec = readFileRuntimeRecord(t, src.runtimePath("rt-1"))
	if !rec.LastHeartbeat.Equal(now) {
		t.Fatalf("heartbeat after update = %v, want %v", rec.LastHeartbeat, now)
	}
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

func TestFileQueueSourceMissingDir(t *testing.T) {
	_, err := NewFileQueueSource("/nonexistent-path-xyz-9999")
	if err == nil {
		t.Fatal("expected error for missing dir")
	}
}

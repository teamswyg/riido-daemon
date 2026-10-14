package controlplane

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestFileQueueSourceSkipsUnavailableProviderForRuntime(t *testing.T) {
	dir := t.TempDir()
	src, err := NewFileQueueSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	registerRuntimeAvailability(t, src, "rt-claude", true, false)
	registerRuntimeAvailability(t, src, "rt-codex", false, true)
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

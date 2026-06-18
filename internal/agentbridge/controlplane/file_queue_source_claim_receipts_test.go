package controlplane

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

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

package controlplane

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestFileQueueSourceReadsTasksFromDir(t *testing.T) {
	dir := t.TempDir()
	src, err := NewFileQueueSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	writeFileQueueSourceTasks(t, dir)

	got1, err := src.ClaimTask(context.Background(), "rt-1")
	if err != nil || got1 == nil {
		t.Fatalf("claim 1: %v / %+v", err, got1)
	}
	got2, _ := src.ClaimTask(context.Background(), "rt-1")
	got3, _ := src.ClaimTask(context.Background(), "rt-1")
	if got1.ID == got2.ID {
		t.Fatalf("claim should not return same task twice")
	}
	if got3 != nil {
		t.Fatalf("expected empty, got %+v", got3)
	}
	assertFileQueueClaimsConsumed(t, dir)
}

func writeFileQueueSourceTasks(t *testing.T, dir string) {
	t.Helper()
	for i, provider := range []string{"claude", "codex"} {
		req := bridge.TaskRequest{ID: "f-" + strconv.Itoa(i), Provider: bridge.Provider(provider), Prompt: "x"}
		body, _ := json.Marshal(req)
		path := filepath.Join(dir, req.ID+".json")
		if err := os.WriteFile(path, body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func assertFileQueueClaimsConsumed(t *testing.T, dir string) {
	t.Helper()
	if remaining := countTopLevelJSON(t, dir); remaining != 0 {
		t.Fatalf("file source should consume top-level tasks, got %d remaining", remaining)
	}
	claims := readClaimRecords(t, filepath.Join(dir, "claims"))
	if len(claims) != 2 {
		t.Fatalf("claim receipts = %+v", claims)
	}
	for _, rec := range claims {
		if rec.SchemaVersion != FileClaimRecordSchemaVersion || rec.RuntimeID != "rt-1" || rec.TaskID == "" {
			t.Fatalf("claim receipt mismatch: %+v", rec)
		}
	}
}

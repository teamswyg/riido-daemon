package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func writeFileQueueTask(t *testing.T, queueDir string) {
	t.Helper()
	req := bridge.TaskRequest{
		ID:       "task-1",
		Provider: bridge.Provider("claude"),
		Prompt:   "hello",
		Metadata: map[string]string{"workspace_id": "workspace-1"},
	}
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(queueDir, "task-1.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertOneReportFile(t *testing.T, reportDir string) {
	t.Helper()
	entries, err := os.ReadDir(reportDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one report file, got %d", len(entries))
	}
}

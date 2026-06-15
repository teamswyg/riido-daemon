package controlplane

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestFileReporterWritesJSONL(t *testing.T) {
	dir := t.TempDir()
	rep, err := NewFileReporter(dir)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	rep.now = func() time.Time { return now }

	if err := rep.StartTask(context.Background(), "task-1"); err != nil {
		t.Fatal(err)
	}
	if err := rep.ReportEvent(context.Background(), "task-1", agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "hello"}); err != nil {
		t.Fatal(err)
	}
	if err := rep.CompleteTask(context.Background(), "task-1", agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one report file, got %d", len(entries))
	}
	body, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	var records []FileReportRecord
	for {
		var rec FileReportRecord
		err := dec.Decode(&rec)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		records = append(records, rec)
	}
	if len(records) != 3 {
		t.Fatalf("records: %+v", records)
	}
	if records[0].Type != "started" || records[0].TaskID != "task-1" {
		t.Fatalf("started record: %+v", records[0])
	}
	if records[1].Event == nil || records[1].Event.Text != "hello" {
		t.Fatalf("event record: %+v", records[1])
	}
	if records[2].Result == nil || records[2].Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("result record: %+v", records[2])
	}
}

func TestFileReporterCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "reports")
	if _, err := NewFileReporter(dir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatalf("report path is not a dir: %s", dir)
	}
}

// --- FileQueueSource ---

func TestFileQueueSourceReadsTasksFromDir(t *testing.T) {
	dir := t.TempDir()
	src, err := NewFileQueueSource(dir)
	if err != nil {
		t.Fatal(err)
	}

	for i, p := range []string{"claude", "codex"} {
		req := bridge.TaskRequest{ID: "f-" + strconv.Itoa(i), Provider: bridge.Provider(p), Prompt: "x"}
		body, _ := json.Marshal(req)
		path := filepath.Join(dir, req.ID+".json")
		if err := os.WriteFile(path, body, 0o644); err != nil {
			t.Fatal(err)
		}
	}

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

	// Tasks must be removed from the top-level dir after claim (no replay).
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

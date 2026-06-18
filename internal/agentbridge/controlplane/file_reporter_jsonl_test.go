package controlplane

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestFileReporterWritesJSONL(t *testing.T) {
	dir := t.TempDir()
	rep, err := NewFileReporter(dir)
	if err != nil {
		t.Fatal(err)
	}
	rep.now = func() time.Time {
		return time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	}

	if err := rep.StartTask(context.Background(), "task-1"); err != nil {
		t.Fatal(err)
	}
	event := agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "hello"}
	if err := rep.ReportEvent(context.Background(), "task-1", event); err != nil {
		t.Fatal(err)
	}
	result := agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}
	if err := rep.CompleteTask(context.Background(), "task-1", result); err != nil {
		t.Fatal(err)
	}

	records := readOnlyFileReportRecords(t, dir)
	if len(records) != 3 {
		t.Fatalf("records: %+v", records)
	}
	assertStartedFileReportRecord(t, records[0])
	assertEventFileReportRecord(t, records[1])
	assertResultFileReportRecord(t, records[2])
}

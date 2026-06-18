package controlplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestMemoryReporterRoundTrip(t *testing.T) {
	rep := NewMemoryReporter()
	if err := rep.StartTask(context.Background(), "t-1"); err != nil {
		t.Fatal(err)
	}
	event := agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: "hi"}
	if err := rep.ReportEvent(context.Background(), "t-1", event); err != nil {
		t.Fatal(err)
	}
	result := agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "ok"}
	if err := rep.CompleteTask(context.Background(), "t-1", result); err != nil {
		t.Fatal(err)
	}
	rec := rep.Recorded("t-1")
	if !rec.Started || len(rec.Events) != 1 {
		t.Fatalf("record events: %+v", rec)
	}
	if rec.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("record result: %+v", rec.Result)
	}
}

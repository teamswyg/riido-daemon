package ingest

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
)

func TestAppendCompletesAndValidatesEnvelope(t *testing.T) {
	sink := &memorySink{}
	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	ingestor, err := New(daemonTestConfig(sink, now))
	if err != nil {
		t.Fatal(err)
	}

	ev, err := ingestor.Append(context.Background(), validNativeConfigDraft())
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	if ev.EventID == "" || ev.EventSchemaVersion != EventSchemaVersionV1 || ev.OccurredAt != now {
		t.Fatalf("completed envelope mismatch: %+v", ev)
	}
	if ev.ActorKind != ir.ActorDaemon || ev.ActorID != "daemon-1" {
		t.Fatalf("actor attribution mismatch: %+v", ev)
	}
	if len(sink.events) != 1 || sink.events[0].EventID != ev.EventID {
		t.Fatalf("sink events: %+v", sink.events)
	}
	assertSinkBatch(t, sink, 1)
}

package ingest

import (
	"context"
	"testing"
)

func TestAppendRedactsSecretsAndAppendsAuditEvent(t *testing.T) {
	sink := &memorySink{}
	ingestor, err := New(agentRedactionTestConfig(sink))
	if err != nil {
		t.Fatal(err)
	}

	ev, err := ingestor.Append(context.Background(), redactionDraft())
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	if len(sink.events) != 2 {
		t.Fatalf("sink events = %d, want audit + redacted event: %+v", len(sink.events), sink.events)
	}
	assertSinkBatch(t, sink, 2)
	assertRedactedEvent(t, sink.events[0], ev.EventID)
	assertRedactionAuditEvent(t, sink.events[1], ev.EventID)
}

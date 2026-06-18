package workdir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
)

func TestRunEventSinkAppendsJSONL(t *testing.T) {
	_, ws := preparedTestWorkspace(t, "run-1")
	sink, err := NewRunEventSink(ws)
	if err != nil {
		t.Fatal(err)
	}
	if err := sink.AppendEvents(context.Background(), []ir.CanonicalEvent{
		testCanonicalEvent("event-1"), testCanonicalEvent("event-2"),
	}); err != nil {
		t.Fatalf("AppendEvents: %v", err)
	}
	body, err := os.ReadFile(sink.Path())
	if err != nil {
		t.Fatalf("read event log: %v", err)
	}
	assertEventLog(t, body)
}

func assertEventLog(t *testing.T, body []byte) {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(body))
	count := 0
	for {
		var got ir.CanonicalEvent
		err := dec.Decode(&got)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		wantID := fmt.Sprintf("event-%d", count+1)
		if got.EventID != wantID || got.Type != ir.EventTaskCreated {
			t.Fatalf("event mismatch: %+v", got)
		}
		count++
	}
	if count != 2 {
		t.Fatalf("event count = %d, want 2", count)
	}
}

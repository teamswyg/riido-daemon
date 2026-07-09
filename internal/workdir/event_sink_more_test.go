package workdir

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
)

func TestRunEventSinkRejectsMissingIRDir(t *testing.T) {
	if _, err := NewRunEventSink(Workspace{}); err == nil {
		t.Fatal("expected missing IR directory error")
	}
}

func TestRunEventSinkAppendEventWritesSingleEvent(t *testing.T) {
	_, ws := preparedTestWorkspace(t, "run-single")
	sink, err := NewRunEventSink(ws)
	if err != nil {
		t.Fatal(err)
	}
	if err := sink.AppendEvent(context.Background(), testCanonicalEvent("event-1")); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(sink.Path())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"EventID":"event-1"`) {
		t.Fatalf("missing event in log: %s", body)
	}
}

func TestRunEventSinkAppendEventsEmptyDoesNotCreateFile(t *testing.T) {
	_, ws := preparedTestWorkspace(t, "run-empty")
	sink, err := NewRunEventSink(ws)
	if err != nil {
		t.Fatal(err)
	}
	if err := sink.AppendEvents(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(sink.Path()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("empty append should not create file, got %v", err)
	}
}

func TestRunEventSinkAppendEventsHonorsCanceledContext(t *testing.T) {
	_, ws := preparedTestWorkspace(t, "run-canceled")
	sink, err := NewRunEventSink(ws)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sink.AppendEvents(ctx, []ir.CanonicalEvent{testCanonicalEvent("event-1")}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if _, err := os.Stat(sink.Path()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("canceled append should not create file, got %v", err)
	}
}

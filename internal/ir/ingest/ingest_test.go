package ingest

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
)

type memorySink struct {
	events      []ir.CanonicalEvent
	appendCalls int
	batchSizes  []int
}

func (s *memorySink) AppendEvent(_ context.Context, ev ir.CanonicalEvent) error {
	s.events = append(s.events, ev)
	return nil
}

func (s *memorySink) AppendEvents(_ context.Context, events []ir.CanonicalEvent) error {
	s.appendCalls++
	s.batchSizes = append(s.batchSizes, len(events))
	s.events = append(s.events, events...)
	return nil
}

func TestAppendCompletesAndValidatesEnvelope(t *testing.T) {
	sink := &memorySink{}
	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	ingestor, err := New(Config{
		Sink:                sink,
		RiidoDaemonVersion:  "riido-daemon.test.v1",
		PolicyBundleVersion: "policy-bundle.test.v1",
		ActorKind:           ir.ActorDaemon,
		ActorID:             "daemon-1",
		Now:                 func() time.Time { return now },
		NewEventID: func(time.Time) (string, error) {
			return "018f0000-0000-7000-8000-000000000001", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ev, err := ingestor.Append(context.Background(), Draft{
		Scope:                 ir.EventScopeRun,
		Type:                  ir.EventNativeConfigInjected,
		TaskID:                "task-1",
		RunID:                 "run-1",
		RuntimeID:             "runtime-1",
		CapabilityFingerprint: "cap-fp-1",
		ProviderKind:          "codex",
		ProtocolKind:          "codex-app-server",
		ProviderVersion:       "codex 1.0",
		AdapterID:             "codex",
		AdapterVersion:        "riido-agentbridge-adapter.v1",
		ProtocolVersion:       "v1",
		NativeConfigVersion:   "nc-1",
		Payload: map[string]any{
			"files": []string{"AGENTS.md"},
		},
	})
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
	if sink.appendCalls != 1 || !containsInt(sink.batchSizes, 1) {
		t.Fatalf("sink batches: calls=%d sizes=%v", sink.appendCalls, sink.batchSizes)
	}
}

func TestAppendRejectsInvalidEnvelopeBeforeSink(t *testing.T) {
	sink := &memorySink{}
	ingestor, err := New(Config{
		Sink:                sink,
		RiidoDaemonVersion:  "riido-daemon.test.v1",
		PolicyBundleVersion: "policy-bundle.test.v1",
		ActorKind:           ir.ActorDaemon,
		ActorID:             "daemon-1",
		NewEventID: func(time.Time) (string, error) {
			return "018f0000-0000-7000-8000-000000000001", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = ingestor.Append(context.Background(), Draft{
		Scope:                 ir.EventScopeRun,
		Type:                  ir.EventNativeConfigInjected,
		TaskID:                "task-1",
		RunID:                 "run-1",
		RuntimeID:             "runtime-1",
		CapabilityFingerprint: "cap-fp-1",
		ProviderKind:          "codex",
		ProtocolKind:          "codex-app-server",
		ProviderVersion:       "codex 1.0",
		AdapterID:             "codex",
		AdapterVersion:        "riido-agentbridge-adapter.v1",
		ProtocolVersion:       "v1",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var envelopeErr EnvelopeError
	if !errors.As(err, &envelopeErr) {
		t.Fatalf("expected EnvelopeError, got %T %v", err, err)
	}
	if !strings.Contains(err.Error(), "NativeConfigVersion") {
		t.Fatalf("error should mention missing NCV: %v", err)
	}
	if len(sink.events) != 0 {
		t.Fatalf("invalid event must not reach sink: %+v", sink.events)
	}
	if sink.appendCalls != 0 {
		t.Fatalf("invalid event must not call sink: calls=%d", sink.appendCalls)
	}
}

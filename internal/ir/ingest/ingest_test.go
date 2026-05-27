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

func TestAppendRedactsSecretsAndAppendsAuditEvent(t *testing.T) {
	sink := &memorySink{}
	ingestor, err := New(Config{
		Sink:                sink,
		RiidoDaemonVersion:  "riido-daemon.test.v1",
		PolicyBundleVersion: "policy-bundle.test.v1",
		ActorKind:           ir.ActorAgent,
		ActorID:             "run-1",
		NewEventID: sequentialEventIDs(
			"018f0000-0000-7000-8000-000000000101",
			"018f0000-0000-7000-8000-000000000102",
		),
	})
	if err != nil {
		t.Fatal(err)
	}

	ev, err := ingestor.Append(context.Background(), Draft{
		Scope:                 ir.EventScopeRun,
		Type:                  ir.EventTextDelta,
		TaskID:                "task-1",
		RunID:                 "run-1",
		RuntimeID:             "runtime-1",
		CapabilityFingerprint: "cap-fp-1",
		ProviderKind:          "claude",
		ProtocolKind:          "claude-jsonl",
		ProviderVersion:       "claude 1.0",
		AdapterID:             "claude",
		AdapterVersion:        "riido-agentbridge-adapter.v1",
		ProtocolVersion:       "v1",
		NativeConfigVersion:   "nc-1",
		Payload: map[string]any{
			"text": "token ghp_" + strings.Repeat("a", 20),
			"nested": map[string]any{
				"url": "https://user:pass@example.com/path",
			},
		},
		Unknown: map[string]any{
			"raw": "RIIDO_TOKEN=" + strings.Repeat("b", 12),
		},
	})
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	if len(sink.events) != 2 {
		t.Fatalf("sink events = %d, want audit + redacted event: %+v", len(sink.events), sink.events)
	}
	if sink.appendCalls != 1 || !containsInt(sink.batchSizes, 2) {
		t.Fatalf("redacted event and audit must share one batch: calls=%d sizes=%v", sink.appendCalls, sink.batchSizes)
	}
	redacted := sink.events[0]
	if redacted.EventID != ev.EventID {
		t.Fatalf("returned event must be redacted event: %+v vs %+v", ev, redacted)
	}
	if strings.Contains(redacted.Payload["text"].(string), "ghp_") {
		t.Fatalf("payload leaked raw github token: %+v", redacted.Payload)
	}
	if strings.Contains(redacted.Unknown["raw"].(string), "RIIDO_TOKEN=") {
		t.Fatalf("unknown leaked raw env token: %+v", redacted.Unknown)
	}
	nested := redacted.Payload["nested"].(map[string]any)
	if strings.Contains(nested["url"].(string), "user:pass") {
		t.Fatalf("nested payload leaked basic auth URL: %+v", nested)
	}

	audit := sink.events[1]
	if audit.Type != ir.EventPolicyViolationDetected {
		t.Fatalf("second event must be audit event: %+v", audit)
	}
	if audit.ActorKind != ir.ActorAgent || audit.ActorID != "run-1" {
		t.Fatalf("audit attribution mismatch: %+v", audit)
	}
	if audit.Payload["category"] != "SECRET_LEAK_ATTEMPTED" || audit.Payload["severity"] != "high" {
		t.Fatalf("audit payload mismatch: %+v", audit.Payload)
	}
	if audit.Payload["sourceEventID"] != ev.EventID || audit.Payload["sourceEventType"] != string(ir.EventTextDelta) {
		t.Fatalf("audit source mismatch: %+v", audit.Payload)
	}
	subject, _ := audit.Payload["subject"].(string)
	for _, want := range []string{"basic-auth-url", "env-secret-assignment", "github-token"} {
		if !strings.Contains(subject, want) {
			t.Fatalf("audit subject %q missing %q", subject, want)
		}
	}
	redactedFields, ok := audit.Payload["redactedFields"].([]string)
	if !ok {
		t.Fatalf("redactedFields type = %T", audit.Payload["redactedFields"])
	}
	for _, want := range []string{"payload.text", "payload.nested.url", "unknown.raw"} {
		if !containsString(redactedFields, want) {
			t.Fatalf("redacted fields %v missing %q", redactedFields, want)
		}
	}

}

func TestNewUUID7EventIDShape(t *testing.T) {
	id, err := NewUUID7EventID(time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 36 || id[14] != '7' {
		t.Fatalf("uuid7 shape mismatch: %q", id)
	}
}

func sequentialEventIDs(ids ...string) func(time.Time) (string, error) {
	next := 0
	return func(time.Time) (string, error) {
		if next >= len(ids) {
			return "", errors.New("no event id left")
		}
		id := ids[next]
		next++
		return id, nil
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsInt(values []int, want int) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

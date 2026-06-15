package claude

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildProviderInputApprovalResponse(t *testing.T) {
	body, err := BuildProviderInput(agentbridge.Command{
		Kind:              agentbridge.CommandApproveTool,
		ToolID:            "tu_1",
		ProviderRequestID: "r_1",
	})
	if err != nil {
		t.Fatalf("BuildProviderInput approve: %v", err)
	}
	raw := string(body)
	for _, want := range []string{`"type":"control_response"`, `"request_id":"r_1"`, `"behavior":"allow"`, `"updatedInput":{}`} {
		if !strings.Contains(raw, want) {
			t.Fatalf("approval response missing %s: %s", want, raw)
		}
	}
}

func TestBuildProviderInputDenyResponse(t *testing.T) {
	body, err := BuildProviderInput(agentbridge.Command{
		Kind:              agentbridge.CommandRejectTool,
		ProviderRequestID: "r_2",
		Reason:            "No shell access",
	})
	if err != nil {
		t.Fatalf("BuildProviderInput deny: %v", err)
	}
	raw := string(body)
	for _, want := range []string{`"request_id":"r_2"`, `"behavior":"deny"`, `"message":"No shell access"`} {
		if !strings.Contains(raw, want) {
			t.Fatalf("deny response missing %s: %s", want, raw)
		}
	}
}

// rate_limit_event is informational (not terminal): a clear Warning, never a
// generic "unknown event" Log and never an EventResult.
func TestTranslateRateLimitEventIsWarning(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"rate_limit_event","rate_limit":{"status":"rejected","resets_at":"2026-06-08T14:00:00Z"}}`)
	events := translate(t, raw)
	if len(events) != 1 {
		t.Fatalf("rate_limit_event: want 1 event, got %d: %+v", len(events), events)
	}
	ev := events[0]
	if ev.Kind != agentbridge.EventWarning {
		t.Fatalf("rate_limit_event: want EventWarning, got %+v", ev)
	}
	if ev.Err == "" {
		t.Fatalf("rate_limit_event: want non-empty detail, got %+v", ev)
	}
}

func TestBuildProviderInputRequiresProviderRequestID(t *testing.T) {
	if _, err := BuildProviderInput(agentbridge.Command{Kind: agentbridge.CommandApproveTool, ToolID: "tu_1"}); err == nil {
		t.Fatal("expected missing provider request id to fail")
	}
}

// log event → Log
func TestTranslateLogEvent(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"log","message":"hello"}`)
	events := translate(t, raw)
	if len(events) != 1 || events[0].Kind != agentbridge.EventLog {
		t.Fatalf("log: %+v", events)
	}
}

// Stderr lines come through as Log events with Source preserved.
func TestTranslateStderr(t *testing.T) {
	raw := agentbridge.RawEvent{
		Source: agentbridge.RawSourceStderr,
		Type:   "stderr",
		Bytes:  []byte("warning thing"),
	}
	events := translate(t, raw)
	if len(events) != 1 || events[0].Kind != agentbridge.EventLog {
		t.Fatalf("stderr: %+v", events)
	}
}

// Malformed lines produce a Warning event so the watchdog can see something
// happened without taking the run to a terminal state.
func TestTranslateMalformedProducesWarning(t *testing.T) {
	raw := agentbridge.RawEvent{
		Source: agentbridge.RawSourceStdout,
		Type:   "malformed",
		Bytes:  []byte("not json"),
	}
	events := translate(t, raw)
	if len(events) != 1 || events[0].Kind != agentbridge.EventWarning {
		t.Fatalf("malformed: %+v", events)
	}
}

// Unknown event types are surfaced as Log, never silently dropped.
func TestTranslateUnknownTypeIsLogged(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"some_new_event_type","foo":"bar"}`)
	events := translate(t, raw)
	if len(events) != 1 || events[0].Kind != agentbridge.EventLog {
		t.Fatalf("unknown type: %+v", events)
	}
}

package claude

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateResultSuccess(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"result","subtype":"success","result":"done","usage":{"input_tokens":3,"output_tokens":7}}`)
	events := translate(t, raw)

	if len(events) == 0 {
		t.Fatalf("expected events")
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult ||
		last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", last)
	}
	if last.Result.Output != "done" {
		t.Fatalf("output: %q", last.Result.Output)
	}

	sawUsage := false
	for _, ev := range events {
		if ev.Kind == agentbridge.EventUsageDelta {
			sawUsage = true
			if ev.Usage.PromptTokens != 3 || ev.Usage.CompletionTokens != 7 {
				t.Fatalf("usage tokens: %+v", ev.Usage)
			}
		}
	}
	if !sawUsage {
		t.Fatalf("expected usage delta in events: %+v", events)
	}
}

func TestTranslateResultError(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"result","subtype":"error","error":"boom"}`)
	events := translate(t, raw)
	last := events[len(events)-1]

	if last.Kind != agentbridge.EventResult ||
		last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("result err: %+v", last)
	}
	if last.Result.Error != "boom" {
		t.Fatalf("error: %q", last.Result.Error)
	}
}

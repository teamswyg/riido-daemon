package openclaw

import (
	"encoding/json"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestParserPrettyFullJSONResult(t *testing.T) {
	p := NewParser()
	r, err := p.FeedStdout([]byte("{\n  \"payloads\": [\n    {\"text\": \"ok\"}\n  ],\n  \"meta\": {\"agentMeta\": {\"sessionId\": \"sess-1\"}}\n}\n"))
	if err != nil {
		t.Fatalf("Feed: %v", err)
	}
	closed, _ := p.Close()
	r = append(r, closed...)
	if len(r) != 1 {
		t.Fatalf("want 1 raw, got %d: %+v", len(r), r)
	}
	if r[0].Type != "full_result" {
		t.Fatalf("type: %q", r[0].Type)
	}
}

func TestParserNDJSONFallback(t *testing.T) {
	p := NewParser()
	chunk := `{"event":"text","text":"chunk1"}` + "\n" + `{"event":"text","text":"chunk2"}` + "\n"
	r, _ := p.FeedStdout([]byte(chunk))
	closed, _ := p.Close()
	r = append(r, closed...)
	if len(r) != 2 {
		t.Fatalf("want 2 raws, got %d", len(r))
	}
	if r[0].Type != "ndjson:text" || r[1].Type != "ndjson:text" {
		t.Fatalf("types: %q %q", r[0].Type, r[1].Type)
	}
}

// --- Translator ---

func tx(t *testing.T, raw agentbridge.RawEvent) []agentbridge.Event {
	t.Helper()
	evs, _, err := Translate(raw)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	return evs
}

func rawFull(t *testing.T, s string) agentbridge.RawEvent {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("fixture: %v", err)
	}
	return agentbridge.RawEvent{Source: agentbridge.RawSourceStdout, Type: "full_result", Payload: m}
}

func TestTranslateFullResultSuccess(t *testing.T) {
	raw := rawFull(t, `{"session_id":"sess-1","text":"hello world","usage":{"prompt_tokens":3,"completion_tokens":7}}`)
	evs := tx(t, raw)
	if len(evs) == 0 {
		t.Fatal("no events")
	}
	last := evs[len(evs)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", last)
	}
	sawSession, sawUsage, sawText := false, false, false
	for _, ev := range evs {
		switch ev.Kind {
		case agentbridge.EventSessionIdentified:
			if ev.SessionID == "sess-1" {
				sawSession = true
			}
		case agentbridge.EventUsageDelta:
			if ev.Usage.PromptTokens == 3 && ev.Usage.CompletionTokens == 7 {
				sawUsage = true
			}
		case agentbridge.EventTextDelta:
			if ev.Text == "hello world" {
				sawText = true
			}
		}
	}
	if !sawSession || !sawUsage || !sawText {
		t.Fatalf("missing session/usage/text in events: %+v", evs)
	}
}

func TestTranslateCurrentFullResultShape(t *testing.T) {
	raw := rawFull(t, `{
		"payloads":[{"text":"ok","mediaUrl":null}],
		"meta":{
			"agentMeta":{
				"sessionId":"integration-openclaw",
				"usage":{"input":14886,"output":2,"total":14888},
				"lastCallUsage":{"input":14886,"output":2,"cacheRead":0,"cacheWrite":0,"total":14888}
			},
			"aborted":false
		}
	}`)
	evs := tx(t, raw)
	var saw struct {
		session, usage, text, result bool
	}
	for _, ev := range evs {
		switch ev.Kind {
		case agentbridge.EventSessionIdentified:
			saw.session = ev.SessionID == "integration-openclaw"
		case agentbridge.EventUsageDelta:
			saw.usage = ev.Usage.PromptTokens == 14886 && ev.Usage.CompletionTokens == 2
		case agentbridge.EventTextDelta:
			saw.text = ev.Text == "ok"
		case agentbridge.EventResult:
			saw.result = ev.Result.Status == agentbridge.ResultCompleted && ev.Result.Output == "ok"
		}
	}
	if !saw.session || !saw.usage || !saw.text || !saw.result {
		t.Fatalf("current full_result shape coverage gap: %+v events=%+v", saw, evs)
	}
}

func TestTranslateFullResultError(t *testing.T) {
	raw := rawFull(t, `{"error":"model rejected"}`)
	evs := tx(t, raw)
	last := evs[len(evs)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("err: %+v", last)
	}
}

func TestTranslateFullResultWithoutTextFailsClosed(t *testing.T) {
	raw := rawFull(t, `{
		"payloads":[],
		"meta":{
			"agentMeta":{
				"sessionId":"integration-openclaw",
				"usage":{"input":14886,"output":0,"total":14886}
			},
			"aborted":false
		}
	}`)
	evs := tx(t, raw)
	last := evs[len(evs)-1]
	if last.Kind != agentbridge.EventResult ||
		last.Result.Status != agentbridge.ResultFailed ||
		last.Result.Error == "" {
		t.Fatalf("empty full_result must fail closed: %+v", last)
	}
}

func TestTranslateNDJSONText(t *testing.T) {
	raw := agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    "ndjson:text",
		Payload: map[string]any{"event": "text", "text": "chunk"},
	}
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventTextDelta || evs[0].Text != "chunk" {
		t.Fatalf("ndjson text: %+v", evs)
	}
}

func TestTranslateMalformedWarning(t *testing.T) {
	evs := tx(t, agentbridge.RawEvent{Source: agentbridge.RawSourceStdout, Type: "malformed", Bytes: []byte("x")})
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventWarning {
		t.Fatalf("malformed: %+v", evs)
	}
}

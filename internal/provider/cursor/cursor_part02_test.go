package cursor

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestParserStripsStdoutStderrPrefixes(t *testing.T) {
	p := NewParser()
	chunk := `stdout: {"type":"text","text":"hi"}` + "\n"
	r, _ := p.FeedStdout([]byte(chunk))
	if len(r) != 1 || r[0].Type != "text" {
		t.Fatalf("stdout prefix not stripped: %+v", r)
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

func rawJSON(t *testing.T, s string) agentbridge.RawEvent {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("fixture: %v", err)
	}
	typ, _ := m["type"].(string)
	return agentbridge.RawEvent{Source: agentbridge.RawSourceStdout, Type: typ, Payload: m}
}

func TestTranslateSystemInit(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"system","subtype":"init","session_id":"sess-1"}`))
	if len(evs) < 2 || evs[0].Kind != agentbridge.EventSessionIdentified {
		t.Fatalf("system: %+v", evs)
	}
}

func TestTranslateText(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"text","text":"hello"}`))
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventTextDelta || evs[0].Text != "hello" {
		t.Fatalf("text: %+v", evs)
	}
}

func TestTranslateAssistantText(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"assistant","content":[{"type":"output_text","text":"x"}]}`))
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventTextDelta || evs[0].Text != "x" {
		t.Fatalf("assistant text: %+v", evs)
	}
}

func TestTranslateAssistantThinking(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"assistant","content":[{"type":"thinking","text":"hmm"}]}`))
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventThinkingDelta {
		t.Fatalf("thinking: %+v", evs)
	}
}

func TestTranslateToolUse(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"go test ./...","password":"raw","note":"sk-ant-`+strings.Repeat("a", 24)+`"}}`))
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventToolCallStarted {
		t.Fatalf("tool_use: %+v", evs)
	}
	if evs[0].Tool.Args["command"] != "go test ./..." {
		t.Fatalf("tool args: %+v", evs[0].Tool.Args)
	}
	if evs[0].Tool.Args["password"] != "[redacted]" {
		t.Fatalf("sensitive args must be redacted: %+v", evs[0].Tool.Args)
	}
	if evs[0].Tool.Args["note"] != "[redacted]" {
		t.Fatalf("secret-looking value must be redacted: %+v", evs[0].Tool.Args)
	}
}

func TestTranslateToolResult(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"tool_result","tool_use_id":"t1"}`))
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventToolCallCompleted {
		t.Fatalf("tool_result: %+v", evs)
	}
}

func TestTranslateResultSuccess(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"result","subtype":"success","result":"done","usage":{"input_tokens":1,"output_tokens":2}}`))
	last := evs[len(evs)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", last)
	}
}

func TestTranslateStepFinishUsageFallback(t *testing.T) {
	evs := tx(t, rawJSON(t, `{"type":"step_finish","usage":{"input_tokens":3,"output_tokens":4}}`))
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventUsageDelta {
		t.Fatalf("usage fallback: %+v", evs)
	}
}

func TestTranslateMalformedWarning(t *testing.T) {
	evs := tx(t, agentbridge.RawEvent{Source: agentbridge.RawSourceStdout, Type: "malformed", Bytes: []byte("x")})
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventWarning {
		t.Fatalf("malformed: %+v", evs)
	}
}

package openclaw

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// --- Command builder ---

func TestBuildStartShape(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Cwd:    "/tmp/work",
		Prompt: "do the thing",
	}, StartOptions{SessionID: "sess-1"})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	for _, want := range []string{"agent", "--local", "--json", "--session-id sess-1", "--message do the thing"} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing %q in %q", want, args)
		}
	}
	if cmd.Dir != "/tmp/work" {
		t.Fatalf("Dir: %q", cmd.Dir)
	}
}

func TestBuildStartUsesRuntimeSelectedExecutable(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Executable: "/opt/riido/bin/openclaw-supported",
		TaskID:     "task-openclaw-1",
		Prompt:     "do the thing",
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Executable != "/opt/riido/bin/openclaw-supported" {
		t.Fatalf("executable = %q", cmd.Executable)
	}
}

func TestBuildStartSystemPromptInlineFallback(t *testing.T) {
	// OpenClaw versions without --system-prompt: the adapter inlines the
	// system prompt into the message so behavior is preserved
	// (spec §3.3).
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Prompt:       "user task",
		SystemPrompt: "be careful",
	}, StartOptions{SessionID: "sess-2"})
	msgFlag := false
	for i, a := range cmd.Args {
		if a == "--message" && i+1 < len(cmd.Args) {
			if !strings.Contains(cmd.Args[i+1], "be careful") || !strings.Contains(cmd.Args[i+1], "user task") {
				t.Fatalf("inline fallback missing content: %q", cmd.Args[i+1])
			}
			msgFlag = true
		}
	}
	if !msgFlag {
		t.Fatalf("--message not built: %v", cmd.Args)
	}
}

func TestBuildStartBlockedArgs(t *testing.T) {
	for _, want := range []string{"--local", "--json", "--session-id", "--message", "--model", "--system-prompt"} {
		if !slices.Contains(BlockedArgs(), want) {
			t.Fatalf("BlockedArgs missing %q: %v", want, BlockedArgs())
		}
	}
	cmd, _ := BuildStart(agentbridge.StartRequest{
		CustomArgs: []string{"--json", "compact", "--my-flag"},
	}, StartOptions{SessionID: "x"})
	if !slices.Contains(cmd.DroppedArgs, "--json") {
		t.Fatalf("--json must be dropped: %v", cmd.DroppedArgs)
	}
	if !strings.Contains(strings.Join(cmd.Args, " "), "--my-flag") {
		t.Fatalf("non-blocked arg lost")
	}
}

func TestBuildStartRequiresSessionID(t *testing.T) {
	_, err := BuildStart(agentbridge.StartRequest{}, StartOptions{})
	if err == nil {
		t.Fatal("expected error without session id")
	}
}

func TestBuildStartDerivesSessionIDFromTaskID(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		TaskID: "task-openclaw-1",
		Prompt: "do the thing",
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "--session-id task-openclaw-1") {
		t.Fatalf("session id not derived from task id: %q", args)
	}
}

func TestBuildStartPrefersResumeSessionID(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		TaskID:          "task-openclaw-1",
		ResumeSessionID: "sess-existing",
		Prompt:          "continue",
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "--session-id sess-existing") {
		t.Fatalf("resume session id not preferred: %q", args)
	}
}

// --- Parser ---

func TestParserFullJSONResult(t *testing.T) {
	p := NewParser()
	r, err := p.FeedStdout([]byte(`{"session_id":"sess-1","text":"hello","usage":{"prompt_tokens":3,"completion_tokens":7}}`))
	if err != nil {
		t.Fatalf("Feed: %v", err)
	}
	closed, _ := p.Close()
	r = append(r, closed...)
	if len(r) != 1 {
		t.Fatalf("want 1 raw, got %d", len(r))
	}
	if r[0].Type != "full_result" {
		t.Fatalf("type: %q", r[0].Type)
	}
}

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
	if !(saw.session && saw.usage && saw.text && saw.result) {
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

package cursor

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

// --- Command ---

// Default profile is RootPrint (cursor-agent -p ...), NOT the legacy
// `chat` subcommand. Current cursor-agent CLI takes -p at the root
// level and treats `chat` as prompt text.
func TestBuildStartDefaultProfileIsRootPrint(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Cwd:    "/tmp/work",
		Prompt: "do the thing",
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	// Required: -p prompt, --output-format stream-json, --workspace cwd.
	for _, want := range []string{"-p do the thing", "--output-format stream-json", "--workspace /tmp/work"} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing %q in %q", want, args)
		}
	}
	// FORBIDDEN by default: bare `chat` subcommand (would be treated as
	// prompt text on current cursor-agent).
	if cmd.Args[0] == "chat" {
		t.Fatalf("default profile must NOT use legacy `chat` subcommand: %v", cmd.Args)
	}
	if cmd.Dir != "/tmp/work" {
		t.Fatalf("Dir: %q", cmd.Dir)
	}
}

// Explicit profile selection must be honored.
func TestBuildStartProfileAgentSubcommand(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work", Prompt: "hi",
	}, StartOptions{Profile: ProfileAgentSubcommand})
	if err != nil {
		t.Fatal(err)
	}
	if len(cmd.Args) == 0 || cmd.Args[0] != "agent" {
		t.Fatalf("agent-subcommand profile must start with 'agent', got %v", cmd.Args)
	}
}

// Legacy chat profile is opt-in only.
func TestBuildStartProfileLegacyChatOptIn(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work", Prompt: "hi",
	}, StartOptions{Profile: ProfileLegacyChat})
	if err != nil {
		t.Fatal(err)
	}
	if len(cmd.Args) == 0 || cmd.Args[0] != "chat" {
		t.Fatalf("legacy-chat profile must start with 'chat', got %v", cmd.Args)
	}
}

// Unknown profile rejected — no silent fallback.
func TestBuildStartProfileUnknownRejected(t *testing.T) {
	_, err := BuildStart(agentbridge.StartRequest{Prompt: "x"}, StartOptions{Profile: "ghost"})
	if err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestBuildStartYoloIsExplicitOptIn(t *testing.T) {
	// Default: NO --yolo.
	cmd, _ := BuildStart(agentbridge.StartRequest{Prompt: "x"}, StartOptions{})
	if strings.Contains(strings.Join(cmd.Args, " "), "--yolo") {
		t.Fatalf("--yolo must NOT be set by default: %v", cmd.Args)
	}
	// AllowYolo still needs an isolated trust tier and explicit policy-bundle allow.
	if _, err := BuildStart(agentbridge.StartRequest{Prompt: "x"}, StartOptions{AllowYolo: true}); err == nil {
		t.Fatal("AllowYolo without policy allow must be rejected")
	}
	if _, err := BuildStart(agentbridge.StartRequest{Prompt: "x"}, StartOptions{
		AllowYolo:           true,
		TrustTier:           policy.TrustTierHost,
		UnsafeBypassAllowed: true,
	}); err == nil {
		t.Fatal("AllowYolo on Host trust tier must be rejected")
	}
	cmd, err := BuildStart(agentbridge.StartRequest{Prompt: "x"}, StartOptions{
		AllowYolo:           true,
		TrustTier:           policy.TrustTierEphemeralVM,
		UnsafeBypassAllowed: true,
	})
	if err != nil {
		t.Fatalf("isolated policy-approved AllowYolo should pass: %v", err)
	}
	if !strings.Contains(strings.Join(cmd.Args, " "), "--yolo") {
		t.Fatalf("--yolo must be set when AllowYolo: %v", cmd.Args)
	}
}

func TestBuildStartBlockedArgs(t *testing.T) {
	for _, want := range []string{"-p", "--output-format", "--yolo"} {
		if !slices.Contains(BlockedArgs(), want) {
			t.Fatalf("BlockedArgs missing %q: %v", want, BlockedArgs())
		}
	}
}

// Cursor doesn't support system prompt / max turns. The adapter must
// surface a Warning event on BuildStart rather than silently dropping
// (spec §3.4).
func TestBuildStartUnsupportedSystemPromptSurfaceWarning(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Prompt:       "x",
		SystemPrompt: "be careful",
		MaxTurns:     5,
	}, StartOptions{})
	if len(cmd.DroppedArgs) == 0 {
		t.Fatalf("expected DroppedArgs to record unsupported features, got none")
	}
	joined := strings.Join(cmd.DroppedArgs, " ")
	if !strings.Contains(joined, "system_prompt") {
		t.Fatalf("system_prompt not surfaced: %v", cmd.DroppedArgs)
	}
	if !strings.Contains(joined, "max_turns") {
		t.Fatalf("max_turns not surfaced: %v", cmd.DroppedArgs)
	}
}

// --- Parser ---

func TestParserStreamJSON(t *testing.T) {
	p := NewParser()
	chunk := `{"type":"system","subtype":"init","session_id":"sess-1"}` + "\n"
	r, _ := p.FeedStdout([]byte(chunk))
	if len(r) != 1 || r[0].Type != "system" {
		t.Fatalf("system: %+v", r)
	}
}

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

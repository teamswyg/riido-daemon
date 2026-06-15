package openclaw

import (
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

func TestBuildStartDerivesProviderSafeSessionIDFromRiidoComponentID(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		TaskID: "-4ckNAErFPZoB721KhZgt",
		Prompt: "do the thing",
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var sessionID string
	for i, arg := range cmd.Args {
		if arg == "--session-id" && i+1 < len(cmd.Args) {
			sessionID = cmd.Args[i+1]
			break
		}
	}
	if sessionID == "" {
		t.Fatalf("session id not found: %v", cmd.Args)
	}
	if strings.HasPrefix(sessionID, "-") {
		t.Fatalf("session id must not start with hyphen: %q", sessionID)
	}
	if !strings.HasPrefix(sessionID, "riido-4ckNAErFPZoB721KhZgt-") {
		t.Fatalf("session id did not preserve task id slug: %q", sessionID)
	}
	if len(sessionID) > 80 {
		t.Fatalf("session id too long: %d %q", len(sessionID), sessionID)
	}
	if !isOpenClawSessionID(sessionID) {
		t.Fatalf("session id is not provider-safe: %q", sessionID)
	}
	if got := sessionIDFromTaskID("-4ckNAErFPZoB721KhZgt"); got != sessionID {
		t.Fatalf("session id must be deterministic: %q != %q", got, sessionID)
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

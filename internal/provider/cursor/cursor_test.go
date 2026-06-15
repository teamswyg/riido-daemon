package cursor

import (
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
	// Required: -p prompt, --output-format stream-json, --workspace cwd,
	// and --trust so headless Cursor does not stop at the workspace trust prompt.
	for _, want := range []string{"-p do the thing", "--output-format stream-json", "--workspace /tmp/work", "--trust"} {
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

func TestBuildStartTrustsDaemonWorkspaceWithoutYolo(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Cwd:    "/tmp/work",
		Prompt: "do the thing",
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "--trust") {
		t.Fatalf("headless workspace must be trusted explicitly: %v", cmd.Args)
	}
	if strings.Contains(args, "--yolo") {
		t.Fatalf("--trust must not imply unsafe --yolo: %v", cmd.Args)
	}
}

func TestBuildStartUsesRuntimeSelectedExecutable(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Executable: "/opt/riido/bin/cursor-selected",
		Prompt:     "do the thing",
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Executable != "/opt/riido/bin/cursor-selected" {
		t.Fatalf("runtime-selected executable lost: %q", cmd.Executable)
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

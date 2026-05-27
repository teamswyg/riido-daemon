package claude

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestBuildStartProtocolCriticalArgs(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
	}, StartOptions{
		PermissionMode: PermissionModeApproval,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	args := strings.Join(cmd.Args, " ")
	for _, want := range []string{
		"-p",
		"--output-format stream-json",
		"--input-format stream-json",
		"--verbose",
		"--permission-mode default",
	} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing protocol-critical arg %q in %q", want, args)
		}
	}
	if cmd.Dir != "/tmp/work" {
		t.Fatalf("Dir not propagated: %q", cmd.Dir)
	}
	if cmd.StdinMode != agentbridge.StdinPipe {
		t.Fatalf("expected StdinPipe, got %q", cmd.StdinMode)
	}
}

func TestBuildStartRequiresExplicitPermissionMode(t *testing.T) {
	_, err := BuildStart(agentbridge.StartRequest{}, StartOptions{})
	if err == nil {
		t.Fatal("expected error when PermissionMode is empty (no implicit bypass)")
	}
}

func TestBuildStartBypassRequiresPolicyAllow(t *testing.T) {
	if _, err := BuildStart(agentbridge.StartRequest{}, StartOptions{
		PermissionMode: PermissionModeBypassDangerous,
	}); err == nil {
		t.Fatal("bypassPermissions without trust tier and policy allow must be rejected")
	}
	if _, err := BuildStart(agentbridge.StartRequest{}, StartOptions{
		PermissionMode:      PermissionModeBypassDangerous,
		TrustTier:           policy.TrustTierHost,
		UnsafeBypassAllowed: true,
	}); err == nil {
		t.Fatal("bypassPermissions on Host trust tier must be rejected")
	}
	cmd, err := BuildStart(agentbridge.StartRequest{}, StartOptions{
		PermissionMode:      PermissionModeBypassDangerous,
		TrustTier:           policy.TrustTierIsolatedContainer,
		UnsafeBypassAllowed: true,
	})
	if err != nil {
		t.Fatalf("isolated policy-approved bypassPermissions should pass: %v", err)
	}
	if !strings.Contains(strings.Join(cmd.Args, " "), "--permission-mode bypassPermissions") {
		t.Fatalf("bypassPermissions mode missing from args: %v", cmd.Args)
	}
}

func TestBuildStartDropsBlockedCustomArgs(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		CustomArgs: []string{
			"--output-format", "text",
			"--permission-mode=bypassPermissions",
			"--my-flag", "ok",
		},
	}, StartOptions{PermissionMode: PermissionModeApproval})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")

	// The caller's --output-format must NOT win.
	if strings.HasSuffix(args, "--output-format text") || strings.Contains(args, "--output-format text ") {
		t.Fatalf("caller overrode --output-format: %q", args)
	}
	// The caller's --permission-mode=bypassPermissions must not bleed through.
	if strings.Contains(args, "--permission-mode=bypassPermissions") {
		t.Fatalf("caller's bypassPermissions bled through: %q", args)
	}
	// Non-blocked custom args survive.
	if !strings.Contains(args, "--my-flag ok") {
		t.Fatalf("non-blocked custom arg lost: %q", args)
	}
	// Dropped args are surfaced.
	if len(cmd.DroppedArgs) == 0 {
		t.Fatalf("expected dropped args to be surfaced, got none")
	}
	for _, want := range []string{"--output-format", "--permission-mode=bypassPermissions"} {
		if !slices.Contains(cmd.DroppedArgs, want) {
			t.Fatalf("expected %q in DroppedArgs %v", want, cmd.DroppedArgs)
		}
	}
}

func TestBuildStartMCPGuard(t *testing.T) {
	// Without --mcp-config path, --strict-mcp-config must NOT be set.
	bare, _ := BuildStart(agentbridge.StartRequest{}, StartOptions{PermissionMode: PermissionModeApproval})
	if strings.Contains(strings.Join(bare.Args, " "), "--strict-mcp-config") {
		t.Fatalf("--strict-mcp-config must not be set without --mcp-config: %v", bare.Args)
	}

	with, _ := BuildStart(agentbridge.StartRequest{}, StartOptions{
		PermissionMode: PermissionModeApproval,
		MCPConfigPath:  "/tmp/mcp.json",
	})
	a := strings.Join(with.Args, " ")
	if !strings.Contains(a, "--strict-mcp-config") || !strings.Contains(a, "--mcp-config /tmp/mcp.json") {
		t.Fatalf("missing strict-mcp-config or --mcp-config when path provided: %q", a)
	}
	if !slices.Contains(with.TempFiles, "/tmp/mcp.json") {
		t.Fatalf("MCP config path should be registered as a temp file for cleanup, got TempFiles=%v", with.TempFiles)
	}
}

func TestBuildStartResumeMaxTurnsModelSystem(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Model:           "claude-opus-4-7",
		SystemPrompt:    "be concise",
		MaxTurns:        4,
		ResumeSessionID: "sess-abc",
	}, StartOptions{PermissionMode: PermissionModeApproval})
	args := strings.Join(cmd.Args, " ")
	for _, want := range []string{
		"--model claude-opus-4-7",
		"--append-system-prompt be concise",
		"--max-turns 4",
		"--resume sess-abc",
	} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing arg %q in %q", want, args)
		}
	}
}

func TestBuildStartExecutableOverride(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{}, StartOptions{
		Executable:     "/opt/anthropic/claude",
		PermissionMode: PermissionModeApproval,
	})
	if cmd.Executable != "/opt/anthropic/claude" {
		t.Fatalf("executable override lost: %q", cmd.Executable)
	}
}

func TestBlockedArgsCoverProtocolCritical(t *testing.T) {
	want := []string{"-p", "--output-format", "--input-format", "--permission-mode", "--mcp-config", "--strict-mcp-config"}
	got := BlockedArgs()
	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Fatalf("BlockedArgs missing %q (got %v)", w, got)
		}
	}
}

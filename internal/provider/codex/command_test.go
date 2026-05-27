package codex

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartProtocolCriticalArgs(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{Cwd: "/tmp/work"}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	for _, want := range []string{"app-server", "--listen", "stdio://"} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing protocol-critical token %q in %q", want, args)
		}
	}
	if cmd.Dir != "/tmp/work" {
		t.Fatalf("Dir not propagated: %q", cmd.Dir)
	}
	if cmd.StdinMode != agentbridge.StdinPipe {
		t.Fatalf("expected StdinPipe, got %q", cmd.StdinMode)
	}
}

func TestBuildStartBlocksListenOverride(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		CustomArgs: []string{"--listen", "tcp://0.0.0.0:9999", "--my-flag"},
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(cmd.DroppedArgs, "--listen") {
		t.Fatalf("--listen must be dropped: %v", cmd.DroppedArgs)
	}
	args := strings.Join(cmd.Args, " ")
	if strings.Contains(args, "tcp://") {
		t.Fatalf("caller's --listen value bled through: %q", args)
	}
	if !strings.Contains(args, "--my-flag") {
		t.Fatalf("non-blocked arg lost: %q", args)
	}
}

func TestBuildStartBlocksUnsafeBypassCustomArgs(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		CustomArgs: []string{
			"--yolo",
			"--dangerously-bypass-approvals-and-sandbox",
			"--sandbox", "danger-full-access",
			"--keep",
			"--sandbox=workspace-write",
		},
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"--yolo",
		"--dangerously-bypass-approvals-and-sandbox",
		"--sandbox",
		"danger-full-access",
	} {
		if !slices.Contains(cmd.DroppedArgs, want) {
			t.Fatalf("unsafe bypass arg %q must be dropped: %v", want, cmd.DroppedArgs)
		}
	}
	args := strings.Join(cmd.Args, " ")
	for _, bad := range []string{
		"--yolo",
		"--dangerously-bypass-approvals-and-sandbox",
		"danger-full-access",
	} {
		if strings.Contains(args, bad) {
			t.Fatalf("unsafe bypass arg %q bled through: %q", bad, args)
		}
	}
	for _, want := range []string{"--keep", "--sandbox=workspace-write"} {
		if !strings.Contains(args, want) {
			t.Fatalf("non-dangerous custom arg %q lost: %q", want, args)
		}
	}
}

func TestBuildStartBlocksCodexDangerSandboxEqualsArg(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		CustomArgs: []string{"--sandbox=danger-full-access", "--safe"},
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(cmd.DroppedArgs, "--sandbox=danger-full-access") {
		t.Fatalf("danger-full-access sandbox arg must be dropped: %v", cmd.DroppedArgs)
	}
	args := strings.Join(cmd.Args, " ")
	if strings.Contains(args, "danger-full-access") {
		t.Fatalf("danger-full-access sandbox bled through: %q", args)
	}
	if !strings.Contains(args, "--safe") {
		t.Fatalf("non-dangerous custom arg lost: %q", args)
	}
}

func TestBuildStartBlocksCodexUnsafeBypassBooleanEqualsArgs(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		CustomArgs: []string{
			"--yolo=true",
			"--dangerously-bypass-approvals-and-sandbox=true",
			"--keep",
		},
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"--yolo=true",
		"--dangerously-bypass-approvals-and-sandbox=true",
	} {
		if !slices.Contains(cmd.DroppedArgs, want) {
			t.Fatalf("unsafe bypass equals-form arg %q must be dropped: %v", want, cmd.DroppedArgs)
		}
	}
	args := strings.Join(cmd.Args, " ")
	for _, bad := range []string{"--yolo", "--dangerously-bypass-approvals-and-sandbox"} {
		if strings.Contains(args, bad) {
			t.Fatalf("unsafe bypass equals-form arg %q bled through: %q", bad, args)
		}
	}
	if !strings.Contains(args, "--keep") {
		t.Fatalf("non-dangerous custom arg lost: %q", args)
	}
}

func TestBuildStartCodexHomeIsolation(t *testing.T) {
	// CODEX_HOME isolation: when StartOptions.CodexHome is set, the env
	// must include CODEX_HOME=<that path>. This prevents Codex from
	// reading the user's global ~/.codex.
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Env: map[string]string{"FOO": "bar"},
	}, StartOptions{CodexHome: "/tmp/codex-task-1"})
	envJoined := strings.Join(cmd.Env, " ")
	if !strings.Contains(envJoined, "CODEX_HOME=/tmp/codex-task-1") {
		t.Fatalf("CODEX_HOME not injected: %v", cmd.Env)
	}
	if !strings.Contains(envJoined, "FOO=bar") {
		t.Fatalf("caller env lost: %v", cmd.Env)
	}
}

func TestBuildStartCallerCannotOverrideCodexHome(t *testing.T) {
	// Caller's CODEX_HOME via req.Env must NOT override the adapter's
	// isolation value.
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Env: map[string]string{"CODEX_HOME": "/home/user/.codex"},
	}, StartOptions{CodexHome: "/tmp/isolated"})
	for _, env := range cmd.Env {
		if env == "CODEX_HOME=/home/user/.codex" {
			t.Fatalf("caller overrode CODEX_HOME: %v", cmd.Env)
		}
	}
	if !slices.Contains(cmd.Env, "CODEX_HOME=/tmp/isolated") {
		t.Fatalf("isolated CODEX_HOME not set: %v", cmd.Env)
	}
}

func TestBuildStartExecutableOverride(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{}, StartOptions{Executable: "/opt/codex/bin/codex"})
	if cmd.Executable != "/opt/codex/bin/codex" {
		t.Fatalf("exe override lost: %q", cmd.Executable)
	}
}

func TestBlockedArgsCoverProtocolCritical(t *testing.T) {
	for _, want := range []string{"--listen"} {
		if !slices.Contains(BlockedArgs(), want) {
			t.Fatalf("BlockedArgs missing %q: %v", want, BlockedArgs())
		}
	}
}

func TestUnsafeBypassArgsCoverSecuritySSOTSurfaces(t *testing.T) {
	for _, want := range []string{
		"--yolo",
		"--dangerously-bypass-approvals-and-sandbox",
		"--sandbox=danger-full-access",
	} {
		if !slices.Contains(UnsafeBypassArgs(), want) {
			t.Fatalf("UnsafeBypassArgs missing %q: %v", want, UnsafeBypassArgs())
		}
	}
}

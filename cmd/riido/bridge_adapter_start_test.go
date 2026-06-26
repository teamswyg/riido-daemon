package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/claude"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
	"github.com/teamswyg/riido-daemon/internal/provider/cursor"
	"github.com/teamswyg/riido-daemon/internal/provider/openclaw"
)

func TestRegisteredAdaptersBuildStartForDaemonRuntime(t *testing.T) {
	for _, adapter := range builtinAgentAdapters() {
		cmd, err := adapter.BuildStart(agentbridge.StartRequest{
			TaskID: "task-" + adapter.Name(),
			Prompt: "do the thing",
			Cwd:    "/tmp/work",
		})
		if err != nil {
			t.Fatalf("%s BuildStart: %v", adapter.Name(), err)
		}
		if cmd.Executable == "" {
			t.Fatalf("%s executable empty", adapter.Name())
		}
		assertBridgeAdapterStart(t, adapter.Name(), cmd)
	}
}

func TestDaemonRegisteredClaudeUsesBetaFullAccess(t *testing.T) {
	cmd, err := builtinDaemonAgentAdapters("/tmp/agentd.sock")[0].BuildStart(agentbridge.StartRequest{
		TaskID: "task-claude",
		Prompt: "do the thing",
		Cwd:    "/tmp/work",
	})
	if err != nil {
		t.Fatalf("daemon claude BuildStart: %v", err)
	}
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "--permission-mode bypassPermissions") {
		t.Fatalf("daemon claude must use beta full-access mode, got %q", args)
	}
}

func assertBridgeAdapterStart(t *testing.T, name string, cmd agentbridge.StartCommand) {
	t.Helper()
	args := strings.Join(cmd.Args, " ")
	switch name {
	case claude.Name:
		assertClaudeBridgeStart(t, args)
	case codex.Name:
		assertCodexBridgeStart(t, cmd.Args, args, cmd.Env)
	case openclaw.Name:
		assertBridgeArgPair(t, cmd.Args, "--session-id", "task-openclaw")
	case cursor.Name:
		if strings.Contains(args, "--yolo") {
			t.Fatalf("cursor daemon adapter must not default to --yolo: %v", cmd.Args)
		}
	}
}

func assertClaudeBridgeStart(t *testing.T, args string) {
	t.Helper()
	if !strings.Contains(args, "--permission-mode default") {
		t.Fatalf("claude daemon adapter must use provider approval mode, got %q", args)
	}
	if strings.Contains(args, "bypassPermissions") {
		t.Fatalf("claude daemon adapter must not default to bypassPermissions: %q", args)
	}
	if strings.Contains(args, "acceptEdits") {
		t.Fatalf("claude daemon adapter must not rely on local edit approval shortcuts: %q", args)
	}
}

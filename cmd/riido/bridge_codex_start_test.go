package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
)

func TestCodexDaemonAdapterPreservesConfiguredCodexHomeWithoutPermissionProfile(t *testing.T) {
	cmd, err := bridgeCodexAdapter{}.BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Env: map[string]string{"CODEX_HOME": "/Users/example/.codex"},
	})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	assertCodexBridgeStart(t, cmd.Args, args, cmd.Env)
	for _, bad := range []string{`"/Users/example/.codex"="none"`, "default_permissions", "permissions.riido-task"} {
		if strings.Contains(args, bad) {
			t.Fatalf("codex adapter must not generate permission profile token %q: %q", bad, args)
		}
	}
	if !containsEnv(cmd.Env, "CODEX_HOME=/Users/example/.codex") {
		t.Fatalf("codex adapter should preserve caller CODEX_HOME for app-server auth: %v", cmd.Env)
	}
}

func TestCodexDaemonAdapterDoesNotDeriveDefaultCodexHomeFromHome(t *testing.T) {
	cmd, err := bridgeCodexAdapter{}.BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Env: map[string]string{"HOME": "/Users/example"},
	})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	if strings.Contains(args, "/Users/example/.codex") || strings.Contains(args, "default_permissions") {
		t.Fatalf("codex adapter must not derive auth home permission profile: %q", args)
	}
	for _, env := range cmd.Env {
		if strings.HasPrefix(env, "CODEX_HOME=") {
			t.Fatalf("codex adapter must not invent CODEX_HOME: %v", cmd.Env)
		}
	}
}

func assertCodexBridgeStart(t *testing.T, argList []string, args string, env []string) {
	t.Helper()
	assertBridgeArgPair(t, argList, "--sandbox", codex.FullAccessSandboxMode)
	if strings.Contains(args, "default_permissions") || strings.Contains(args, "permissions.riido-task") {
		t.Fatalf("codex adapter must not generate a task-scoped permission profile: %q", args)
	}
	for _, value := range env {
		if strings.HasPrefix(value, "CODEX_HOME=") && value != "CODEX_HOME=/Users/example/.codex" {
			t.Fatalf("codex adapter must not invent CODEX_HOME: %v", env)
		}
	}
}

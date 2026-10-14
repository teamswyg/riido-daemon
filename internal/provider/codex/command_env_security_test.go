package codex

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartPreservesEnvWithoutPermissionProfile(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Env: map[string]string{"FOO": "bar", "CODEX_HOME": "/Users/example/.codex"},
	}, StartOptions{})
	args := strings.Join(cmd.Args, " ")
	for _, bad := range []string{`"/Users/example/.codex"="none"`, "default_permissions", "permissions.riido-task"} {
		if strings.Contains(args, bad) {
			t.Fatalf("permission profile token %q must not be generated: %q", bad, args)
		}
	}
	for _, want := range []string{"FOO=bar", "CODEX_HOME=/Users/example/.codex"} {
		if !slices.Contains(cmd.Env, want) {
			t.Fatalf("caller env %q lost: %v", want, cmd.Env)
		}
	}
}

func TestBuildStartDoesNotGenerateToolchainPermissionProfile(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Env: map[string]string{
			"HOME":   "/Users/example",
			"GOROOT": "/usr/local/go",
		},
	}, StartOptions{})
	args := strings.Join(cmd.Args, " ")
	for _, bad := range toolchainPermissionProfileTokens() {
		if strings.Contains(args, bad) {
			t.Fatalf("toolchain permission profile token %q must not be generated in %q", bad, args)
		}
	}
	assertArgPair(t, cmd.Args, "--sandbox", FullAccessSandboxMode)
}

func TestBuildStartDoesNotInventCodexHome(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Env: map[string]string{"HOME": "/home/user"},
	}, StartOptions{})
	args := strings.Join(cmd.Args, " ")
	if strings.Contains(args, "/home/user/.codex") || strings.Contains(args, "default_permissions") {
		t.Fatalf("Codex auth home permission profile must not be invented: %q", args)
	}
	for _, env := range cmd.Env {
		if strings.HasPrefix(env, "CODEX_HOME=") {
			t.Fatalf("Codex adapter must not invent CODEX_HOME: %v", cmd.Env)
		}
	}
}

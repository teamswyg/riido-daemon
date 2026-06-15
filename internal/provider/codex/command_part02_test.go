package codex

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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
	for _, bad := range []string{
		`"/usr/local/go"="read"`,
		`"/Users/example/.rustup"="read"`,
		`"/Users/example/.cargo"="read"`,
		`"/Users/example/Library/Caches/go-build"="write"`,
		"default_permissions",
	} {
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

func TestBuildStartExecutableOverride(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{}, StartOptions{Executable: "/opt/codex/bin/codex"})
	if cmd.Executable != "/opt/codex/bin/codex" {
		t.Fatalf("exe override lost: %q", cmd.Executable)
	}
}

func TestBuildStartUsesRuntimeSelectedExecutable(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{Executable: "/opt/riido/bin/codex-selected"}, StartOptions{})
	if cmd.Executable != "/opt/riido/bin/codex-selected" {
		t.Fatalf("runtime-selected executable lost: %q", cmd.Executable)
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
	} {
		if !slices.Contains(UnsafeBypassArgs(), want) {
			t.Fatalf("UnsafeBypassArgs missing %q: %v", want, UnsafeBypassArgs())
		}
	}
}

func TestSandboxOverrideArgsCoverDaemonOwnedSandboxSelection(t *testing.T) {
	for _, want := range []string{"--sandbox", "-s"} {
		if !slices.Contains(SandboxOverrideArgs(), want) {
			t.Fatalf("SandboxOverrideArgs missing %s: %v", want, SandboxOverrideArgs())
		}
	}
}

func assertArgPair(t *testing.T, args []string, key, value string) {
	t.Helper()
	for i := 0; i+1 < len(args); i++ {
		if args[i] == key && args[i+1] == value {
			return
		}
	}
	t.Fatalf("missing arg pair %s %s in %v", key, value, args)
}

func assertArgBefore(t *testing.T, args []string, before, after string) {
	t.Helper()
	beforeIndex := -1
	afterIndex := -1
	for i, arg := range args {
		if arg == before && beforeIndex == -1 {
			beforeIndex = i
		}
		if arg == after && afterIndex == -1 {
			afterIndex = i
		}
	}
	if beforeIndex == -1 || afterIndex == -1 || beforeIndex >= afterIndex {
		t.Fatalf("expected %q before %q in %v", before, after, args)
	}
}

func assertArgCount(t *testing.T, args []string, key string, want int) {
	t.Helper()
	got := 0
	for _, arg := range args {
		if arg == key {
			got++
		}
	}
	if got != want {
		t.Fatalf("arg %q count = %d, want %d in %v", key, got, want, args)
	}
}

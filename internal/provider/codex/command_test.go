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
	for _, want := range []string{"app-server", "--sandbox", FullAccessSandboxMode, "--listen", "stdio://"} {
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
	assertArgPair(t, cmd.Args, "--sandbox", FullAccessSandboxMode)
	for _, bad := range []string{"default_permissions", "permissions.riido-task"} {
		if strings.Contains(args, bad) {
			t.Fatalf("permission profile token %q must not be generated in %q", bad, args)
		}
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
	for _, want := range []string{"--yolo", "--dangerously-bypass-approvals-and-sandbox"} {
		if !slices.Contains(cmd.DroppedArgs, want) {
			t.Fatalf("unsafe bypass arg %q must be dropped: %v", want, cmd.DroppedArgs)
		}
	}
	for _, want := range []string{"--sandbox", "danger-full-access", "--sandbox=workspace-write"} {
		if !slices.Contains(cmd.DroppedArgs, want) {
			t.Fatalf("sandbox override arg %q must be dropped: %v", want, cmd.DroppedArgs)
		}
	}
	args := strings.Join(cmd.Args, " ")
	for _, bad := range []string{"--yolo", "--dangerously-bypass-approvals-and-sandbox", "--sandbox=workspace-write"} {
		if strings.Contains(args, bad) {
			t.Fatalf("unsafe bypass arg %q bled through: %q", bad, args)
		}
	}
	assertArgPair(t, cmd.Args, "--sandbox", FullAccessSandboxMode)
	if !strings.Contains(args, "--keep") {
		t.Fatalf("non-dangerous custom arg lost: %q", args)
	}
}

func TestBuildStartBlocksConfigOverrideArgs(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		CustomArgs: []string{
			"-c", "default_permissions=\"unsafe\"",
			"--config=permissions.unsafe.filesystem={\":minimal\"=\"read\"}",
			"--enable", "experimental_untrusted",
			"--disable=some_guard",
			"--safe",
		},
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"-c",
		"default_permissions=\"unsafe\"",
		"--config=permissions.unsafe.filesystem={\":minimal\"=\"read\"}",
		"--enable",
		"experimental_untrusted",
		"--disable=some_guard",
	} {
		if !slices.Contains(cmd.DroppedArgs, want) {
			t.Fatalf("config override arg %q must be dropped: %v", want, cmd.DroppedArgs)
		}
	}
	args := strings.Join(cmd.Args, " ")
	for _, bad := range []string{"unsafe", "experimental_untrusted", "some_guard"} {
		if strings.Contains(args, bad) {
			t.Fatalf("config override value %q bled through: %q", bad, args)
		}
	}
	if !strings.Contains(args, "--safe") {
		t.Fatalf("non-critical custom arg lost: %q", args)
	}
}

func TestBuildStartBlocksCallerSandboxOverride(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		CustomArgs: []string{"--sandbox=danger-full-access", "--sandbox", "workspace-write", "--safe"},
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"--sandbox=danger-full-access", "--sandbox", "workspace-write"} {
		if !slices.Contains(cmd.DroppedArgs, want) {
			t.Fatalf("sandbox override arg %q must be dropped: %v", want, cmd.DroppedArgs)
		}
	}
	args := strings.Join(cmd.Args, " ")
	if strings.Contains(args, "workspace-write") || strings.Contains(args, "--sandbox=danger-full-access") {
		t.Fatalf("caller sandbox override bled through: %q", args)
	}
	assertArgPair(t, cmd.Args, "--sandbox", FullAccessSandboxMode)
	if !strings.Contains(args, "--safe") {
		t.Fatalf("non-sandbox custom arg lost: %q", args)
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
	if !slices.Contains(SandboxOverrideArgs(), "--sandbox") {
		t.Fatalf("SandboxOverrideArgs missing --sandbox: %v", SandboxOverrideArgs())
	}
}

func assertArgPair(t *testing.T, args []string, key string, value string) {
	t.Helper()
	for i := 0; i+1 < len(args); i++ {
		if args[i] == key && args[i+1] == value {
			return
		}
	}
	t.Fatalf("missing arg pair %s %s in %v", key, value, args)
}

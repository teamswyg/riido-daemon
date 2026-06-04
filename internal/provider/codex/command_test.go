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
	for _, want := range []string{
		`default_permissions="riido-task"`,
		`permissions.riido-task.filesystem={":minimal"="read","/tmp/work"="write"}`,
		`permissions.riido-task.network={enabled=true}`,
	} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing daemon permission profile token %q in %q", want, args)
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

func TestBuildStartDeniesCodexAuthHomeInPermissionProfile(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Env: map[string]string{"FOO": "bar"},
	}, StartOptions{AuthHomeDenyPath: "/Users/example/.codex"})
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, `"/Users/example/.codex"="none"`) {
		t.Fatalf("Codex auth home must be denied in permission profile: %q", args)
	}
	envJoined := strings.Join(cmd.Env, " ")
	if !strings.Contains(envJoined, "FOO=bar") {
		t.Fatalf("caller env lost: %v", cmd.Env)
	}
}

func TestBuildStartAllowsCommonToolchainRootsWithoutDangerSandbox(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Env: map[string]string{
			"HOME":   "/Users/example",
			"GOROOT": "/usr/local/go",
		},
	}, StartOptions{})
	args := strings.Join(cmd.Args, " ")
	for _, want := range []string{
		`"/usr/local/go"="read"`,
		`"/Users/example/.rustup"="read"`,
		`"/Users/example/.cargo"="read"`,
		`"/Users/example/Library/Caches/go-build"="write"`,
	} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing toolchain permission %q in %q", want, args)
		}
	}
	if strings.Contains(args, "danger-full-access") {
		t.Fatalf("toolchain permission must not require danger sandbox: %q", args)
	}
}

func TestBuildStartDerivesCodexAuthHomeDenyPathFromEnv(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Env: map[string]string{"CODEX_HOME": "/home/user/.codex"},
	}, StartOptions{})
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, `"/home/user/.codex"="none"`) {
		t.Fatalf("caller CODEX_HOME must be denied in permission profile: %q", args)
	}
	if !slices.Contains(cmd.Env, "CODEX_HOME=/home/user/.codex") {
		t.Fatalf("caller env should still reach Codex app-server auth process: %v", cmd.Env)
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
		"--sandbox=danger-full-access",
	} {
		if !slices.Contains(UnsafeBypassArgs(), want) {
			t.Fatalf("UnsafeBypassArgs missing %q: %v", want, UnsafeBypassArgs())
		}
	}
}

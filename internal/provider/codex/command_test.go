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
	assertArgBefore(t, cmd.Args, "--sandbox", "app-server")
	assertArgCount(t, cmd.Args, "--sandbox", 1)
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
			"-s", "read-only",
			"--keep",
			"--sandbox=workspace-write",
			"-s=workspace-write",
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
	for _, want := range []string{"--sandbox", "danger-full-access", "-s", "read-only", "--sandbox=workspace-write", "-s=workspace-write"} {
		if !slices.Contains(cmd.DroppedArgs, want) {
			t.Fatalf("sandbox override arg %q must be dropped: %v", want, cmd.DroppedArgs)
		}
	}
	args := strings.Join(cmd.Args, " ")
	for _, bad := range []string{"--yolo", "--dangerously-bypass-approvals-and-sandbox", "--sandbox=workspace-write", "-s=workspace-write", "read-only"} {
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
		CustomArgs: []string{"--sandbox=danger-full-access", "--sandbox", "workspace-write", "-s", "read-only", "-s=workspace-write", "--safe"},
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"--sandbox=danger-full-access", "--sandbox", "workspace-write", "-s", "read-only", "-s=workspace-write"} {
		if !slices.Contains(cmd.DroppedArgs, want) {
			t.Fatalf("sandbox override arg %q must be dropped: %v", want, cmd.DroppedArgs)
		}
	}
	args := strings.Join(cmd.Args, " ")
	if strings.Contains(args, "workspace-write") || strings.Contains(args, "read-only") || strings.Contains(args, "--sandbox=danger-full-access") {
		t.Fatalf("caller sandbox override bled through: %q", args)
	}
	assertArgPair(t, cmd.Args, "--sandbox", FullAccessSandboxMode)
	assertArgCount(t, cmd.Args, "--sandbox", 1)
	if !strings.Contains(args, "--safe") {
		t.Fatalf("non-sandbox custom arg lost: %q", args)
	}
}

func TestBuildStartUsesOnlyDaemonGeneratedSandboxSelection(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		CustomArgs: []string{
			"--sandbox", FullAccessSandboxMode,
			"--sandbox=danger-full-access",
			"-s", FullAccessSandboxMode,
			"--safe",
		},
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertArgPair(t, cmd.Args, "--sandbox", FullAccessSandboxMode)
	assertArgCount(t, cmd.Args, "--sandbox", 1)
	for _, dropped := range []string{"--sandbox", FullAccessSandboxMode, "--sandbox=danger-full-access", "-s"} {
		if !slices.Contains(cmd.DroppedArgs, dropped) {
			t.Fatalf("caller sandbox token %q must be dropped: %v", dropped, cmd.DroppedArgs)
		}
	}
	if !slices.Contains(cmd.Args, "--safe") {
		t.Fatalf("non-sandbox custom arg lost: %v", cmd.Args)
	}
}

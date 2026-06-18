package codex

import (
	"slices"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartBlocksCallerSandboxOverride(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		CustomArgs: []string{"--sandbox=danger-full-access", "--sandbox", "workspace-write", "-s", "read-only", "-s=workspace-write", "--safe"},
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range sandboxOverrideTokens() {
		if !slices.Contains(cmd.DroppedArgs, want) {
			t.Fatalf("sandbox override arg %q must be dropped: %v", want, cmd.DroppedArgs)
		}
	}
	if containsAnyCommandArg(cmd.Args, "workspace-write", "read-only", "--sandbox=danger-full-access") {
		t.Fatalf("caller sandbox override bled through: %v", cmd.Args)
	}
	assertArgPair(t, cmd.Args, "--sandbox", FullAccessSandboxMode)
	assertArgCount(t, cmd.Args, "--sandbox", 1)
	if !slices.Contains(cmd.Args, "--safe") {
		t.Fatalf("non-sandbox custom arg lost: %v", cmd.Args)
	}
}

func TestBuildStartUsesOnlyDaemonGeneratedSandboxSelection(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		CustomArgs: []string{"--sandbox", FullAccessSandboxMode, "--sandbox=danger-full-access", "-s", FullAccessSandboxMode, "--safe"},
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

func sandboxOverrideTokens() []string {
	return []string{"--sandbox=danger-full-access", "--sandbox", "workspace-write", "-s", "read-only", "-s=workspace-write"}
}

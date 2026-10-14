package codex

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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
	for _, want := range unsafeBypassAndSandboxTokens() {
		if !slices.Contains(cmd.DroppedArgs, want) {
			t.Fatalf("unsafe/sandbox override arg %q must be dropped: %v", want, cmd.DroppedArgs)
		}
	}
	if containsAnyCommandArg(cmd.Args, "--yolo", "--dangerously-bypass-approvals-and-sandbox", "--sandbox=workspace-write", "-s=workspace-write", "read-only") {
		t.Fatalf("unsafe bypass arg bled through: %q", strings.Join(cmd.Args, " "))
	}
	assertArgPair(t, cmd.Args, "--sandbox", FullAccessSandboxMode)
	if !slices.Contains(cmd.Args, "--keep") {
		t.Fatalf("non-dangerous custom arg lost: %v", cmd.Args)
	}
}

func unsafeBypassAndSandboxTokens() []string {
	return []string{
		"--yolo",
		"--dangerously-bypass-approvals-and-sandbox",
		"--sandbox",
		"danger-full-access",
		"-s",
		"read-only",
		"--sandbox=workspace-write",
		"-s=workspace-write",
	}
}

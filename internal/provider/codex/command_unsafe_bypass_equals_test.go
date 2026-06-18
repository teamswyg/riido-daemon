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

package openclaw

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartBlockedArgs(t *testing.T) {
	for _, want := range []string{"--local", "--json", "--session-id", "--message", "--model", "--system-prompt"} {
		if !slices.Contains(BlockedArgs(), want) {
			t.Fatalf("BlockedArgs missing %q: %v", want, BlockedArgs())
		}
	}
	cmd, _ := BuildStart(agentbridge.StartRequest{
		CustomArgs: []string{"--json", "compact", "--my-flag"},
	}, StartOptions{SessionID: "x"})
	if !slices.Contains(cmd.DroppedArgs, "--json") {
		t.Fatalf("--json must be dropped: %v", cmd.DroppedArgs)
	}
	if !strings.Contains(strings.Join(cmd.Args, " "), "--my-flag") {
		t.Fatalf("non-blocked arg lost")
	}
}

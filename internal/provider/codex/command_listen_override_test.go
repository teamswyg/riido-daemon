package codex

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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

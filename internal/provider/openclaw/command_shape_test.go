package openclaw

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartShape(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		Cwd:    "/tmp/work",
		Prompt: "do the thing",
	}, StartOptions{SessionID: "sess-1"})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	for _, want := range []string{"agent", "--local", "--json", "--session-id sess-1", "--message do the thing"} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing %q in %q", want, args)
		}
	}
	if cmd.Dir != "/tmp/work" {
		t.Fatalf("Dir: %q", cmd.Dir)
	}
}

package claude

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartProtocolCriticalArgs(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{Cwd: "/tmp/work"}, safeStartOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	args := strings.Join(cmd.Args, " ")
	for _, want := range []string{
		"-p",
		"--output-format stream-json",
		"--input-format stream-json",
		"--verbose",
		"--permission-mode default",
	} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing protocol-critical arg %q in %q", want, args)
		}
	}
	if cmd.Dir != "/tmp/work" {
		t.Fatalf("Dir not propagated: %q", cmd.Dir)
	}
	if cmd.StdinMode != agentbridge.StdinPipe {
		t.Fatalf("expected StdinPipe, got %q", cmd.StdinMode)
	}
}

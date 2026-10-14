package codex

import (
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

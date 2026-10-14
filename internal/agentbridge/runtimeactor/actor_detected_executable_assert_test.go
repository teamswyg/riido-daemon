package runtimeactor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func expectStartRequest(
	t *testing.T,
	startReqCh <-chan agentbridge.StartRequest,
	selected string,
	launchPath string,
) {
	t.Helper()
	select {
	case req := <-startReqCh:
		if req.Executable != selected {
			t.Fatalf("BuildStart request executable = %q, want %q", req.Executable, selected)
		}
		if got := req.Env["PATH"]; got != launchPath {
			t.Fatalf("BuildStart PATH = %q, want %q", got, launchPath)
		}
	case <-time.After(time.Second):
		t.Fatal("BuildStart request not observed")
	}
}

func expectSpawnCommand(
	t *testing.T,
	p *fakeProcess,
	selected string,
	launchPath string,
) {
	t.Helper()
	cmd := p.commandAt(0)
	if got := cmd.Executable; got != selected {
		t.Fatalf("spawn executable = %q, want %q", got, selected)
	}
	if got, ok := envListValue(cmd.Env, "PATH"); !ok || got != launchPath {
		t.Fatalf("spawn PATH = %q ok=%v, want %q; env=%v", got, ok, launchPath, cmd.Env)
	}
}

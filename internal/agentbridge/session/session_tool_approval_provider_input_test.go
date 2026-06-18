package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func assertApprovalProviderInput(t *testing.T, running providerInputReceiver) {
	t.Helper()
	select {
	case got := <-running.StdinRecv():
		if string(got) != "approve:req-1\n" {
			t.Fatalf("provider input = %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("provider input was not written")
	}
}

type providerInputReceiver interface {
	StdinRecv() <-chan []byte
}

func assertApproveToolCommand(t *testing.T, cmd agentbridge.Command) {
	t.Helper()
	if cmd.Kind != agentbridge.CommandApproveTool ||
		cmd.ToolID != "tool-1" ||
		cmd.ProviderRequestID != "req-1" {
		t.Fatalf("unexpected provider input command: %+v", cmd)
	}
}

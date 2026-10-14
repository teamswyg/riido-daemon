package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func assertApprovalProviderInput(t *testing.T, running providerInputReceiver) {
	t.Helper()
	assertProviderInput(t, running, "approve:req-1\n")
}

func assertRejectionProviderInput(t *testing.T, running providerInputReceiver) {
	t.Helper()
	assertProviderInput(t, running, "reject:req-1\n")
}

func assertProviderInput(t *testing.T, running providerInputReceiver, want string) {
	t.Helper()
	select {
	case got := <-running.StdinRecv():
		if string(got) != want {
			t.Fatalf("provider input = %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("provider input was not written")
	}
}

type providerInputReceiver interface {
	StdinRecv() <-chan []byte
}

func assertToolApprovalCommand(t *testing.T, cmd agentbridge.Command, wantKind agentbridge.CommandKind) {
	t.Helper()
	if cmd.Kind != wantKind ||
		cmd.ToolID != "tool-1" ||
		cmd.ProviderRequestID != "req-1" {
		t.Fatalf("unexpected provider input command: %+v", cmd)
	}
}

package session

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type approvalInputAdapter struct {
	burstAdapter
	input []byte
	err   error
}

func (a approvalInputAdapter) BuildProviderInput(agentbridge.Command) ([]byte, error) {
	return a.input, a.err
}

func fillFakeStdin(t *testing.T, proc *process.FakeRunning) {
	t.Helper()
	for i := 0; i < cap(proc.StdinRecv()); i++ {
		if err := proc.WriteStdin([]byte("x")); err != nil {
			t.Fatalf("fill stdin %d: %v", i, err)
		}
	}
}

func assertWarning(t *testing.T, events []agentbridge.Event, text string) {
	t.Helper()
	if len(events) != 1 || events[0].Kind != agentbridge.EventWarning || events[0].Text != text {
		t.Fatalf("warning events = %+v, want %q", events, text)
	}
}

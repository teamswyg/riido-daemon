package codex

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type codexHandshakeFixture struct {
	t       *testing.T
	parser  agentbridge.Parser
	rpc     *RPCActor
	running *process.FakeRunning
}

func newCodexHandshakeFixture(t *testing.T) *codexHandshakeFixture {
	t.Helper()
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	proc, err := fake.Start(context.Background(), process.Command{Executable: "codex"})
	if err != nil {
		t.Fatal(err)
	}
	return &codexHandshakeFixture{
		t:       t,
		parser:  NewParser(),
		rpc:     StartRPCActor(context.Background()),
		running: proc.(*process.FakeRunning),
	}
}

func (f *codexHandshakeFixture) close() {
	f.rpc.Close()
}

package codex

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (f *codexHandshakeFixture) drainUntil(done func(agentbridge.RawEvent, []agentbridge.Event) bool) bool {
	reached := false
	drainHandshake(
		f.t,
		f.parser,
		f.running.Stdout(),
		f.rpc,
		time.Second,
		func(raw agentbridge.RawEvent, evs []agentbridge.Event) bool {
			reached = done(raw, evs)
			return reached
		},
	)
	return reached
}

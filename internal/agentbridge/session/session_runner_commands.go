package session

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (r *sessionRunner) executeCommands(ctx context.Context, cmds []agentbridge.Command) []agentbridge.Event {
	builder, _ := r.cfg.ProtocolDriver.(agentbridge.ProviderInputBuilder)
	return executeCommands(ctx, r.proc, r.cfg.Adapter, cmds, r.cfg.ProcessKillTimeout, builder)
}

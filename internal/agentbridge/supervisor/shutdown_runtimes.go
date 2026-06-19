package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) deregisterRuntimes(ctx context.Context, runtimes []*runtimeactor.Actor) {
	for _, rt := range runtimes {
		status, err := rt.Status(ctx)
		if err != nil || status.RuntimeID == "" {
			continue
		}
		_ = a.cfg.Source.DeregisterRuntime(ctx, status.RuntimeID)
	}
}

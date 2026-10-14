package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) deregisterRuntimes(ctx context.Context, runtimes []*runtimeactor.Actor) {
	pending := runtimeIDsForDeregistration(ctx, runtimes)
	for len(pending) > 0 {
		pending = a.deregisterRuntimeIDs(ctx, pending)
		if len(pending) == 0 || waitShutdownRetry(ctx) {
			return
		}
	}
}

func runtimeIDsForDeregistration(ctx context.Context, runtimes []*runtimeactor.Actor) []string {
	ids := make([]string, 0, len(runtimes))
	for _, rt := range runtimes {
		status, err := rt.Status(ctx)
		if err != nil || status.RuntimeID == "" {
			continue
		}
		ids = append(ids, status.RuntimeID)
	}
	return ids
}

func (a *Actor) deregisterRuntimeIDs(ctx context.Context, ids []string) []string {
	pending := ids[:0]
	for _, runtimeID := range ids {
		if err := a.cfg.Source.DeregisterRuntime(ctx, runtimeID); err != nil {
			pending = append(pending, runtimeID)
		}
	}
	return pending
}

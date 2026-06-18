package supervisor

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func (a *Actor) Start(ctx context.Context) error {
	runtimes := configuredRuntimes(a.cfg)
	for _, rt := range runtimes {
		status, err := rt.Status(ctx)
		if err != nil {
			return fmt.Errorf("supervisor: runtime status: %w", err)
		}
		if err := a.register(ctx, status); err != nil {
			return err
		}
	}
	go a.run(ctx, runtimes)
	return nil
}

func configuredRuntimes(cfg Config) []*runtimeactor.Actor {
	if len(cfg.Runtimes) > 0 {
		out := make([]*runtimeactor.Actor, 0, len(cfg.Runtimes))
		for _, rt := range cfg.Runtimes {
			if rt != nil {
				out = append(out, rt)
			}
		}
		return out
	}
	if cfg.Runtime != nil {
		return []*runtimeactor.Actor{cfg.Runtime}
	}
	return nil
}

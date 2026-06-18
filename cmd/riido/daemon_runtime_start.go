package main

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func startDaemonRuntimeActors(ctx lifecycle.Context, settings daemonSettings, reporter controlplane.TaskReporterPort, log logging.Logger) ([]*runtimeactor.Actor, error) {
	rtActors, err := newDaemonRuntimeActors(settings, builtinAgentAdapters(), daemonToolApprovalResolver(reporter))
	if err != nil {
		return nil, err
	}
	if err := startDaemonRuntimes(ctx, rtActors, log); err != nil {
		return nil, err
	}
	log.Printf("runtimeactors started: %d providers", len(rtActors))
	return rtActors, nil
}

func startDaemonRuntimes(ctx lifecycle.Context, runtimes []*runtimeactor.Actor, log logging.Logger) error {
	started := make([]*runtimeactor.Actor, 0, len(runtimes))
	for _, rt := range runtimes {
		if err := rt.Start(ctx.Context()); err != nil {
			stopDaemonRuntimes(lifecycle.ShutdownForced, started, log)
			return daemonWrapf(ErrDaemonRuntime, "serve.start-runtime", err, "runtimeactor.Start")
		}
		started = append(started, rt)
	}
	return nil
}

func stopDaemonRuntimes(level lifecycle.ShutdownLevel, runtimes []*runtimeactor.Actor, log logging.Logger) {
	shutdownCtx, cancel := lifecycle.DetachedDefaultShutdown(level)
	defer cancel()
	stopRuntimeActors(shutdownCtx, runtimes, log)
}

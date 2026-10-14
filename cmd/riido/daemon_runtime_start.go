package main

import (
	"sync"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func startDaemonRuntimeActors(ctx lifecycle.Context, settings daemonSettings, reporter controlplane.TaskReporterPort, socketPath string, log logging.Logger) ([]*runtimeactor.Actor, error) {
	rtActors, err := newDaemonRuntimeActors(settings, builtinDaemonAgentAdapters(socketPath), daemonToolApprovalResolver(reporter))
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
	errs := make([]error, len(runtimes))
	var wg sync.WaitGroup
	for i, rt := range runtimes {
		wg.Go(func() {
			errs[i] = rt.Start(ctx.Context())
		})
	}
	wg.Wait()
	started := make([]*runtimeactor.Actor, 0, len(runtimes))
	var firstErr error
	for i, err := range errs {
		if err == nil {
			started = append(started, runtimes[i])
		} else if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		stopDaemonRuntimes(lifecycle.ShutdownForced, started, log)
		return daemonWrapf(ErrDaemonRuntime, "serve.start-runtime", firstErr, "runtimeactor.Start")
	}
	return nil
}

func stopDaemonRuntimes(level lifecycle.ShutdownLevel, runtimes []*runtimeactor.Actor, log logging.Logger) {
	shutdownCtx, cancel := lifecycle.DetachedDefaultShutdown(level)
	defer cancel()
	stopRuntimeActors(shutdownCtx, runtimes, log)
}

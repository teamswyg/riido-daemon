package main

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/supervisor"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/workdir"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func startDaemonSupervisor(ctx lifecycle.Context, settings daemonSettings, runtimes []*runtimeactor.Actor, source controlplane.TaskSourcePort, reporter controlplane.TaskReporterPort, workdirAdapter workdir.Adapter) (*supervisor.Actor, error) {
	supActor, err := supervisor.New(supervisor.Config{
		DaemonID:            settings.DaemonID,
		RiidoDaemonVersion:  settings.DaemonVersion,
		Runtimes:            runtimes,
		Source:              source,
		Reporter:            reporter,
		Workdir:             workdirAdapter,
		PollEvery:           settings.PollEvery,
		IdlePollEvery:       settings.IdlePollEvery,
		HeartbeatEvery:      settings.HeartbeatEvery,
		PolicyBundleVersion: settings.PolicyBundle,
		PolicyBundle:        settings.PolicyBundleDoc,
		RuntimeTrustTier:    policy.TrustTierHost,
	})
	if err != nil {
		return nil, daemonWrapf(ErrDaemonSupervisor, "serve.new-supervisor", err, "supervisor.New")
	}
	if err := supActor.Start(ctx.Context()); err != nil {
		return nil, daemonWrapf(ErrDaemonSupervisor, "serve.start-supervisor", err, "supervisor.Start")
	}
	return supActor, nil
}

func stopDaemonSupervisor(level lifecycle.ShutdownLevel, supActor *supervisor.Actor, log logging.Logger) {
	shutdownCtx, cancel := lifecycle.DetachedDefaultShutdown(level)
	defer cancel()
	if err := supActor.StopLifecycle(shutdownCtx); err != nil {
		log.Printf("supervisor stop error level=%s: %v", shutdownCtx.ShutdownLevel(), err)
	}
}

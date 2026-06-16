package main

import (
	"errors"
	"net"
	"os"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/supervisor"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
	"github.com/teamswyg/riido-daemon/internal/workdir"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// serveAgentDaemon opens the Unix socket, constructs one RuntimeActor per
// built-in provider adapter, answers status/health, and returns when ctx is
// canceled or SIGTERM arrives.
func serveAgentDaemon(ctx lifecycle.Context, flags startFlags, settings daemonSettings, log logging.Logger) error {
	// Remove a stale socket from a previous run.
	_ = os.Remove(flags.socket)

	startedAt := time.Now()
	shutdownLevel := lifecycle.NormalizeShutdownLevel(ctx.ShutdownLevel())
	stopPprof, _, err := startDaemonPprofServer(ctx, settings.PprofAddr, log)
	if err != nil {
		return err
	}
	defer stopPprof()

	rtActors, err := newDaemonRuntimeActors(settings, builtinAgentAdapters())
	if err != nil {
		return err
	}
	startedRuntimes := make([]*runtimeactor.Actor, 0, len(rtActors))
	for _, rt := range rtActors {
		if err := rt.Start(ctx.Context()); err != nil {
			shutdownCtx, cancel := lifecycle.DetachedDefaultShutdown(lifecycle.ShutdownForced)
			stopRuntimeActors(shutdownCtx, startedRuntimes, log)
			cancel()
			return daemonWrapf(ErrDaemonRuntime, "serve.start-runtime", err, "runtimeactor.Start")
		}
		startedRuntimes = append(startedRuntimes, rt)
	}
	defer func() {
		shutdownCtx, cancel := lifecycle.DetachedDefaultShutdown(shutdownLevel)
		defer cancel()
		stopRuntimeActors(shutdownCtx, rtActors, log)
	}()

	log.Printf("runtimeactors started: %d providers", len(rtActors))

	taskSource, taskReporter, controlPlaneKind, err := buildDaemonControlPlane(settings, startedAt)
	if err != nil {
		return err
	}
	workdirAdapter := workdir.NewFSAdapter(settings.WorkdirRoot)
	supActor, err := supervisor.New(supervisor.Config{
		DaemonID:            settings.DaemonID,
		RiidoDaemonVersion:  settings.DaemonVersion,
		Runtimes:            rtActors,
		Source:              taskSource,
		Reporter:            taskReporter,
		Workdir:             workdirAdapter,
		PollEvery:           settings.PollEvery,
		IdlePollEvery:       settings.IdlePollEvery,
		HeartbeatEvery:      settings.HeartbeatEvery,
		PolicyBundleVersion: settings.PolicyBundle,
		PolicyBundle:        settings.PolicyBundleDoc,
		RuntimeTrustTier:    policy.TrustTierHost,
	})
	if err != nil {
		return daemonWrapf(ErrDaemonSupervisor, "serve.new-supervisor", err, "supervisor.New")
	}
	if err := supActor.Start(ctx.Context()); err != nil {
		return daemonWrapf(ErrDaemonSupervisor, "serve.start-supervisor", err, "supervisor.Start")
	}
	defer func() {
		shutdownCtx, cancel := lifecycle.DetachedDefaultShutdown(shutdownLevel)
		defer cancel()
		if err := supActor.StopLifecycle(shutdownCtx); err != nil {
			log.Printf("supervisor stop error level=%s: %v", shutdownCtx.ShutdownLevel(), err)
		}
	}()
	log.Printf("supervisor started workdir_root=%s control_plane=%s queue_dir=%s report_dir=%s", settings.WorkdirRoot, controlPlaneKind, settings.TaskQueueDir, settings.TaskReportDir)
	stopCleanup := startWorkdirCleanupLoop(ctx, workdirAdapter, settings, log)
	defer stopCleanup()

	ln, err := net.Listen("unix", flags.socket)
	if err != nil {
		return daemonWrapf(ErrDaemonSocket, "serve.listen", err, "listen %s", flags.socket)
	}
	defer func() {
		_ = ln.Close()
		_ = os.Remove(flags.socket)
	}()

	// Honor ctx cancellation, POSIX signals, AND in-socket "shutdown"
	// requests together. The shutdown channel is buffered so the
	// handler's non-blocking send can never deadlock if multiple
	// clients race a stop request.
	signalCtx, stop := lifecycle.Notify(ctx.WithShutdownLevel(shutdownLevel), daemonInterruptSignals()...)
	defer stop()

	shutdownCh := make(chan lifecycle.ShutdownLevel, 8)
	done := make(chan lifecycle.ShutdownLevel, 1)
	go func() {
		var level lifecycle.ShutdownLevel
		select {
		case <-signalCtx.Done():
			level = lifecycle.NormalizeShutdownLevel(signalCtx.ShutdownLevel())
			log.Printf("daemon shutdown requested level=%s source=signal", level)
		case level = <-shutdownCh:
			level = lifecycle.NormalizeShutdownLevel(level)
			log.Printf("daemon shutdown requested level=%s source=socket", level)
		}
		done <- level
		_ = ln.Close()
	}()

	log.Printf("daemon listening on %s", flags.socket)
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case level := <-done:
				shutdownLevel = level
				log.Printf("daemon shutting down")
				return nil
			default:
			}
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			log.Printf("accept error: %v", err)
			continue
		}
		go handleDaemonConn(conn, flags, settings, startedAt, rtActors, shutdownCh, log)
	}
}

func newDaemonRuntimeActors(settings daemonSettings, adapters []agentbridge.Adapter) ([]*runtimeactor.Actor, error) {
	out := make([]*runtimeactor.Actor, 0, len(adapters))
	for _, adapter := range adapters {
		name := strings.TrimSpace(adapter.Name())
		if name == "" {
			return nil, daemonErrorf(ErrDaemonRuntime, "runtime.new", "runtimeactor.New: adapter name is required")
		}
		rt, err := newDaemonRuntimeActor(settings, providerRuntimeID(settings.DaemonID, name), adapter, settings.RuntimeAgents)
		if err != nil {
			return nil, daemonWrapf(ErrDaemonRuntime, "runtime.new", err, "runtimeactor.New(%s)", name)
		}
		out = append(out, rt)
	}
	if len(out) == 0 {
		return nil, daemonErrorf(ErrDaemonRuntime, "runtime.new", "runtimeactor.New: at least one adapter is required")
	}
	return out, nil
}

func newDaemonRuntimeActor(settings daemonSettings, runtimeID string, adapter agentbridge.Adapter, agents []runtimeactor.AgentStatus) (*runtimeactor.Actor, error) {
	return runtimeactor.New(runtimeactor.Config{
		RuntimeID:           runtimeID,
		Owner:               settings.RuntimeOwner,
		DeviceName:          settings.DeviceName,
		Agents:              agents,
		Models:              daemonRuntimeModels(adapter.Name()),
		Adapters:            []agentbridge.Adapter{adapter},
		Process:             processexec.New(),
		MaxConcurrent:       settings.RuntimeMaxConcurrent,
		AutoApprove:         daemonToolAutoApprover(settings),
		ToolStartGate:       daemonToolStartGate(settings),
		PolicyBundleVersion: settings.PolicyBundle,
	})
}

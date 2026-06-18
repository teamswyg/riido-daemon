package main

import (
	"os"
	"time"

	"github.com/teamswyg/riido-daemon/internal/logging"
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

	taskSource, taskReporter, controlPlaneKind, err := buildDaemonControlPlane(settings, startedAt)
	if err != nil {
		return err
	}

	rtActors, err := startDaemonRuntimeActors(ctx, settings, taskReporter, log)
	if err != nil {
		return err
	}
	defer func() {
		stopDaemonRuntimes(shutdownLevel, rtActors, log)
	}()

	workdirAdapter := workdir.NewFSAdapter(settings.WorkdirRoot)
	supActor, err := startDaemonSupervisor(ctx, settings, rtActors, taskSource, taskReporter, workdirAdapter)
	if err != nil {
		return err
	}
	defer func() {
		stopDaemonSupervisor(shutdownLevel, supActor, log)
	}()
	log.Printf("supervisor started workdir_root=%s control_plane=%s queue_dir=%s report_dir=%s", settings.WorkdirRoot, controlPlaneKind, settings.TaskQueueDir, settings.TaskReportDir)

	stopCleanup := startWorkdirCleanupLoop(ctx, workdirAdapter, settings, log)
	defer stopCleanup()

	level, err := serveDaemonSocket(ctx, flags, settings, startedAt, rtActors, shutdownLevel, log)
	if err != nil {
		return err
	}
	shutdownLevel = level
	log.Printf("daemon shutting down")
	return nil
}

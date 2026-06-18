package main

import (
	"net"
	"os"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func writeStatus(conn net.Conn, flags startFlags, settings daemonSettings, startedAt time.Time, runtimes []*runtimeactor.Actor) {
	obs := observeDaemon(runtimes)
	s := daemonStatus{
		SchemaVersion:  DaemonStatusSchemaVersion,
		DaemonID:       settings.DaemonID,
		DaemonVersion:  settings.DaemonVersion,
		PID:            os.Getpid(),
		UptimeSeconds:  int(time.Since(startedAt).Seconds()),
		Health:         "ok",
		Ready:          obs.ready,
		Readiness:      obs.readyText(),
		Profile:        settings.Profile,
		ServerURL:      settings.ServerURL,
		DeviceName:     settings.DeviceName,
		WorkspaceCount: settings.WorkspaceCount,
		SocketPath:     flags.socket,
		LogFile:        flags.logFile,
		PIDFile:        flags.pidFile,
		RunningTasks:   obs.metrics.RunningTasks,
		Metrics:        obs.metrics,
		Runtimes:       obs.runtimes,
		StartedAt:      startedAt.UTC().Format(time.RFC3339Nano),
	}
	_ = writeDaemonJSON(conn, s)
}

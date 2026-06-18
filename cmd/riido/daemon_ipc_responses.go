package main

import (
	"net"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func writeShutdownAck(conn net.Conn, level lifecycle.ShutdownLevel) {
	_ = writeDaemonJSON(conn, map[string]string{
		"schema_version": DaemonStatusSchemaVersion,
		"shutdown":       "accepted",
		"shutdown_level": level.String(),
	})
}

func writeHealth(conn net.Conn) {
	_ = writeDaemonJSON(conn, map[string]string{
		"schema_version": DaemonStatusSchemaVersion,
		"health":         "ok",
	})
}

func writeReady(conn net.Conn, runtimes []*runtimeactor.Actor) {
	obs := observeDaemon(runtimes)
	_ = writeDaemonJSON(conn, map[string]any{
		"schema_version":     DaemonStatusSchemaVersion,
		"health":             "ok",
		"ready":              obs.ready,
		"readiness":          obs.readyText(),
		"runtime_count":      obs.metrics.RuntimeCount,
		"runtime_responding": obs.metrics.RuntimeResponding,
	})
}

func writeMetrics(conn net.Conn, runtimes []*runtimeactor.Actor) {
	obs := observeDaemon(runtimes)
	_ = writeDaemonJSON(conn, map[string]any{
		"schema_version": DaemonStatusSchemaVersion,
		"metrics":        obs.metrics,
	})
}

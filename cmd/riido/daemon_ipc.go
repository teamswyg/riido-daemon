package main

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// daemonRequest is the JSON envelope read off the socket.
type daemonRequest struct {
	Method        daemonMethod `json:"method"`
	ShutdownLevel string       `json:"shutdown_level,omitempty"`
	Force         bool         `json:"force,omitempty"`
}

func handleDaemonConn(conn net.Conn, flags startFlags, settings daemonSettings, startedAt time.Time, runtimes []*runtimeactor.Actor, shutdownCh chan<- lifecycle.ShutdownLevel, log logging.Logger) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	var req daemonRequest
	dec := json.NewDecoder(conn)
	if err := dec.Decode(&req); err != nil {
		// EOF here means the peer closed without sending a request
		// (typical: the background-start parent's socket-readiness
		// probe). Drop silently — don't try to write a status reply
		// to a closed conn or worse, hand a misleading "" method to
		// the switch below.
		if !errors.Is(err, io.EOF) {
			log.Printf("decode request: %v", err)
		}
		return
	}
	log.Printf("%s request received", req.Method)
	switch req.Method {
	case daemonMethodStatus, daemonMethodDefault:
		writeStatus(conn, flags, settings, startedAt, runtimes)
	case daemonMethodHealth:
		writeHealth(conn)
	case daemonMethodReady:
		writeReady(conn, runtimes)
	case daemonMethodMetrics:
		writeMetrics(conn, runtimes)
	case daemonMethodShutdown:
		level := req.lifecycleShutdownLevel()
		writeShutdownAck(conn, level)
		// Non-blocking signal — repeated shutdown requests are harmless.
		select {
		case shutdownCh <- level:
		default:
		}
		log.Printf("shutdown request received level=%s", level)
	default:
		if err := writeDaemonJSON(conn, map[string]any{"error": "unknown method", "method": string(req.Method)}); err != nil {
			log.Printf("write unknown-method response: %v", err)
		}
	}
}

func writeDaemonJSON(conn net.Conn, value any) error {
	return json.NewEncoder(conn).Encode(value)
}

func (r daemonRequest) lifecycleShutdownLevel() lifecycle.ShutdownLevel {
	if r.Force {
		return lifecycle.ShutdownForced
	}
	if level, ok := lifecycle.ParseShutdownLevel(r.ShutdownLevel); ok {
		return lifecycle.NormalizeShutdownLevel(level)
	}
	return lifecycle.ShutdownGraceful
}

func writeShutdownAck(conn net.Conn, level lifecycle.ShutdownLevel) {
	_ = writeDaemonJSON(conn, map[string]string{
		"schema_version": DaemonStatusSchemaVersion,
		"shutdown":       "accepted",
		"shutdown_level": level.String(),
	})
}

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

type daemonObservation struct {
	runtimes []runtimeactor.Status
	metrics  daemonMetrics
	ready    bool
}

func (o daemonObservation) readyText() string {
	if o.ready {
		return "ready"
	}
	return "not-ready"
}

func observeDaemon(runtimes []*runtimeactor.Actor) daemonObservation {
	ctx, cancel := lifecycle.WithTimeout(lifecycle.Background(), 2*time.Second)
	defer cancel()

	obs := daemonObservation{
		runtimes: make([]runtimeactor.Status, 0, len(runtimes)),
		metrics:  daemonMetrics{RuntimeCount: len(runtimes)},
	}
	for _, rt := range runtimes {
		rtStatus, err := rt.Status(ctx.Context())
		if err != nil {
			continue
		}
		obs.runtimes = append(obs.runtimes, rtStatus)
		obs.metrics.RuntimeResponding++
		obs.metrics.RunningTasks += rtStatus.RunningSessions
		for _, cap := range rtStatus.Capabilities {
			if cap.Available {
				obs.metrics.ProviderAvailable++
			} else {
				obs.metrics.ProviderUnavailable++
			}
		}
	}
	obs.ready = obs.metrics.RuntimeResponding == obs.metrics.RuntimeCount
	return obs
}

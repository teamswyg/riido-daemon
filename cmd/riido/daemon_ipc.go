package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// daemonRequest is the JSON envelope read off the socket.
type daemonRequest struct {
	Method string `json:"method"`
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
	case "status", "":
		writeStatus(conn, flags, settings, startedAt, runtimes)
	case "health":
		writeHealth(conn)
	case "ready":
		writeReady(conn, runtimes)
	case "metrics":
		writeMetrics(conn, runtimes)
	case "shutdown":
		writeShutdownAck(conn)
		// Non-blocking signal — repeated shutdown requests are harmless.
		select {
		case shutdownCh <- lifecycle.ShutdownGraceful:
		default:
		}
		log.Printf("shutdown request received")
	default:
		if err := writeDaemonJSON(conn, map[string]any{"error": "unknown method", "method": req.Method}); err != nil {
			log.Printf("write unknown-method response: %v", err)
		}
	}
}

func writeDaemonJSON(conn net.Conn, value any) error {
	return json.NewEncoder(conn).Encode(value)
}

func writeShutdownAck(conn net.Conn) {
	_ = writeDaemonJSON(conn, map[string]string{
		"schema_version": DaemonStatusSchemaVersion,
		"shutdown":       "accepted",
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

func runDaemonStatus(args []string) error {
	sock, err := requireSocketFlag(args)
	if err != nil {
		return err
	}
	return daemonCall(sock, "status")
}

func runDaemonHealth(args []string) error {
	sock, err := requireSocketFlag(args)
	if err != nil {
		return err
	}
	return daemonCall(sock, "health")
}

func runDaemonReady(args []string) error {
	sock, err := requireSocketFlag(args)
	if err != nil {
		return err
	}
	return daemonCall(sock, "ready")
}

func runDaemonMetrics(args []string) error {
	sock, err := requireSocketFlag(args)
	if err != nil {
		return err
	}
	return daemonCall(sock, "metrics")
}

func requireSocketFlag(args []string) (string, error) {
	for i := 0; i < len(args); i++ {
		if args[i] == "--socket" {
			i++
			if i >= len(args) {
				return "", daemonErrorf(ErrDaemonUsage, "ipc.parse-socket", "--socket requires a path")
			}
			return args[i], nil
		}
	}
	return defaultAgentDaemonSocket()
}

func defaultAgentDaemonSocket() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", daemonWrapf(ErrDaemonIO, "ipc.default-socket.user-home", err, "resolve user home")
	}
	root, err := hostintegration.DefaultAppDataRoot(hostintegration.AppDataRootInput{
		Channel:  hostintegration.DistributionChannelDevLocal,
		HostOS:   hostintegration.HostOSDarwin,
		UserHome: home,
	})
	if err != nil {
		return "", daemonWrapf(ErrDaemonConfig, "ipc.default-socket.app-data-root", err, "resolve default app data root")
	}
	endpoint, err := hostintegration.DefaultLocalIPCEndpoint(hostintegration.LocalIPCEndpointInput{
		Channel:     hostintegration.DistributionChannelDevLocal,
		HostOS:      hostintegration.HostOSDarwin,
		AppDataRoot: root,
		Owner:       hostintegration.LocalIPCOwnerHelper,
		Name:        "agentd",
	})
	if err != nil {
		return "", daemonWrapf(ErrDaemonConfig, "ipc.default-socket.endpoint", err, "resolve default local IPC endpoint")
	}
	return endpoint.Path, nil
}

func defaultDaemonLockPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", daemonWrapf(ErrDaemonIO, "daemon.default-lock.user-home", err, "resolve user home")
	}
	return filepath.Join(home, ".riido", ".lock"), nil
}

func daemonCall(sock, method string) error {
	conn, err := net.DialTimeout("unix", sock, 2*time.Second)
	if err != nil {
		return daemonWrapf(ErrDaemonSocket, "ipc.call.dial", err, "dial %s", sock)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if err := json.NewEncoder(conn).Encode(daemonRequest{Method: method}); err != nil {
		return daemonWrapf(ErrDaemonSocket, "ipc.call.encode", err, "encode request")
	}
	body, err := io.ReadAll(conn)
	if err != nil && !errors.Is(err, io.EOF) {
		return daemonWrapf(ErrDaemonSocket, "ipc.call.read", err, "read response")
	}
	_, err = os.Stdout.Write(body)
	return err
}

func runDaemonStop(args []string) error {
	socket := ""
	pidFile := ""
	timeoutSeconds := 5
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--socket":
			i++
			if i >= len(args) {
				return daemonErrorf(ErrDaemonUsage, "stop.parse-flags", "--socket requires a path")
			}
			socket = args[i]
		case "--pid-file":
			i++
			if i >= len(args) {
				return daemonErrorf(ErrDaemonUsage, "stop.parse-flags", "--pid-file requires a path")
			}
			pidFile = args[i]
		case "--timeout-seconds":
			i++
			if i >= len(args) {
				return daemonErrorf(ErrDaemonUsage, "stop.parse-flags", "--timeout-seconds requires a value")
			}
			v, err := strconv.Atoi(args[i])
			if err != nil || v <= 0 {
				return daemonWrapf(ErrDaemonUsage, "stop.parse-flags", err, "--timeout-seconds must be positive int: %v", args[i])
			}
			timeoutSeconds = v
		case "--help", "-h":
			printUsage()
			return nil
		default:
			return daemonErrorf(ErrDaemonUsage, "stop.parse-flags", "unknown argument: %s", args[i])
		}
	}
	if socket == "" && pidFile == "" {
		return daemonErrorf(ErrDaemonUsage, "stop.parse-flags", "daemon stop requires at least one of --socket or --pid-file")
	}

	timeout := time.Duration(timeoutSeconds) * time.Second

	// 1. Socket shutdown first (preferred — cooperative, no signals).
	if socket != "" {
		if ok := tryShutdownViaSocket(socket, timeout); ok {
			return nil
		}
	}

	// 2. PID SIGTERM fallback.
	if pidFile == "" {
		return daemonErrorf(ErrDaemonSocket, "stop.socket-fallback", "daemon stop: socket %s did not respond and --pid-file is not provided", socket)
	}
	return stopViaPIDFile(pidFile, timeout)
}

// tryShutdownViaSocket sends a `shutdown` request to the daemon's Unix
// socket. Returns true when (a) the request was accepted AND (b) the
// daemon visibly stopped accepting connections within timeout.
//
// A "no daemon at this socket" case (Dial fails immediately) also
// reports true so the operator doesn't see a redundant SIGTERM fallback
// when there's nothing to stop. The caller decides whether to follow up
// with a PID-file fallback.
func tryShutdownViaSocket(socket string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("unix", socket, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.SetDeadline(time.Now().Add(1 * time.Second))
	if err := json.NewEncoder(conn).Encode(daemonRequest{Method: "shutdown"}); err != nil {
		_ = conn.Close()
		return false
	}
	// Drain the ack so the server-side write completes before we close.
	_, _ = io.ReadAll(conn)
	_ = conn.Close()

	// Wait for the daemon to actually stop listening.
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("unix", socket, 100*time.Millisecond)
		if err != nil {
			return true
		}
		_ = c.Close()
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func stopViaPIDFile(pidFile string, timeout time.Duration) error {
	raw, err := os.ReadFile(pidFile)
	if err != nil {
		return daemonWrapf(ErrDaemonIO, "stop.read-pid-file", err, "read pid file")
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		return daemonWrapf(ErrDaemonProcess, "stop.parse-pid", err, "parse pid")
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return daemonWrapf(ErrDaemonProcess, "stop.find-process", err, "find process %d", pid)
	}
	if err := signalDaemonProcessTerm(proc); err != nil {
		return daemonWrapf(ErrDaemonProcess, "stop.terminate", err, "terminate daemon process")
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !daemonProcessExists(proc) {
			return nil // gone
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err := signalDaemonProcessKill(proc); err != nil {
		return daemonWrapf(ErrDaemonProcess, "stop.kill", err, "kill daemon process")
	}
	return nil
}

func runDaemonLogs(args []string) error {
	logFile := ""
	lines := 50
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--log-file":
			i++
			if i >= len(args) {
				return daemonErrorf(ErrDaemonUsage, "logs.parse-flags", "--log-file requires a path")
			}
			logFile = args[i]
		case "--lines":
			i++
			if i >= len(args) {
				return daemonErrorf(ErrDaemonUsage, "logs.parse-flags", "--lines requires a value")
			}
			v, err := strconv.Atoi(args[i])
			if err != nil || v <= 0 {
				return daemonWrapf(ErrDaemonUsage, "logs.parse-flags", err, "--lines must be positive int")
			}
			lines = v
		case "--help", "-h":
			printUsage()
			return nil
		default:
			return daemonErrorf(ErrDaemonUsage, "logs.parse-flags", "unknown argument: %s", args[i])
		}
	}
	if logFile == "" {
		return daemonErrorf(ErrDaemonUsage, "logs.parse-flags", "--log-file is required")
	}
	f, err := os.Open(logFile)
	if err != nil {
		return daemonWrapf(ErrDaemonIO, "logs.open", err, "open log")
	}
	defer f.Close()

	// Simple naive tail: read everything, print the last N lines.
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 4096), 1024*1024)
	var all []string
	for scanner.Scan() {
		all = append(all, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return daemonWrapf(ErrDaemonIO, "logs.scan", err, "scan log")
	}
	from := 0
	if len(all) > lines {
		from = len(all) - lines
	}
	for _, ln := range all[from:] {
		fmt.Println(ln)
	}
	return nil
}

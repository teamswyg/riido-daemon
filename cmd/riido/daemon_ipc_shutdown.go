package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// tryShutdownViaSocket sends a `shutdown` request to the daemon's Unix
// socket. Returns true when (a) the request was accepted AND (b) the
// daemon visibly stopped accepting connections within timeout.
//
// A "no daemon at this socket" case (Dial fails immediately) also
// reports true so the operator doesn't see a redundant SIGTERM fallback
// when there's nothing to stop. The caller decides whether to follow up
// with a PID-file fallback.
func tryShutdownViaSocket(socket string, timeout time.Duration, level lifecycle.ShutdownLevel) bool {
	conn, err := net.DialTimeout("unix", socket, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.SetDeadline(time.Now().Add(1 * time.Second))
	if err := json.NewEncoder(conn).Encode(daemonRequest{
		Method:        daemonMethodShutdown,
		ShutdownLevel: level.String(),
		Force:         level.IsForced(),
	}); err != nil {
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
	if err := validateDaemonProcessIdentity(pid); err != nil {
		return err
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

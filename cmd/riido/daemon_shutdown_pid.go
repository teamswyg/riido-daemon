package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func stopViaPIDFile(pidFile string, timeout time.Duration, level lifecycle.ShutdownLevel) error {
	pid, err := parseDaemonPIDFile(pidFile)
	if err != nil {
		return err
	}
	identity, ok, err := loadDaemonPIDIdentity(pidFile)
	if err != nil {
		return daemonWrapf(ErrDaemonIO, "stop.read-pid-identity", err, "read pid identity")
	}
	if !ok {
		return daemonErrorf(ErrDaemonProcess, "stop.verify-pid", "daemon pid identity %s is missing; refusing PID fallback", daemonPIDIdentityPath(pidFile))
	}
	if identity.Socket != "" && tryShutdownViaSocket(identity.Socket, timeout, level) {
		return nil
	}
	if err := validateDaemonProcessIdentity(pid, identity); err != nil {
		return err
	}
	return stopDaemonProcessByPID(pid, timeout)
}

func parseDaemonPIDFile(pidFile string) (int, error) {
	raw, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, daemonWrapf(ErrDaemonIO, "stop.read-pid-file", err, "read pid file")
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		return 0, daemonWrapf(ErrDaemonProcess, "stop.parse-pid", err, "parse pid")
	}
	return pid, nil
}

func stopDaemonProcessByPID(pid int, timeout time.Duration) error {
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
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err := signalDaemonProcessKill(proc); err != nil {
		return daemonWrapf(ErrDaemonProcess, "stop.kill", err, "kill daemon process")
	}
	return nil
}

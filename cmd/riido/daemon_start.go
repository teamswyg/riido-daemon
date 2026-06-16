package main

import (
	"context"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	c9lock "github.com/teamswyg/riido-daemon/internal/lock"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// runDaemonStartForeground is the in-process daemon — it spawns the
// RuntimeActor, opens the socket, and serves until ctx is cancelled or
// SIGTERM/SIGINT/shutdown-request fires. The background wrapper
// re-invokes the same binary with --foreground to land in this path.
func runDaemonStartForeground(ctx lifecycle.Context, flags startFlags) error {
	settings, err := loadDaemonSettings()
	if err != nil {
		return err
	}
	lock, err := c9lock.AcquireFile(ctx.Context(), flags.lockFile)
	if err != nil {
		return daemonWrapf(ErrDaemonLock, "start.acquire-lock", err, "acquire daemon singleton lock %s", flags.lockFile)
	}
	defer func() {
		if releaseErr := lock.Release(); releaseErr != nil {
			_, _ = os.Stderr.WriteString("riido daemon: release lock: " + releaseErr.Error() + "\n")
		}
	}()

	logSink, closeLog, err := openLogSink(flags.logFile)
	if err != nil {
		return daemonWrapf(ErrDaemonIO, "start.open-log", err, "open log sink")
	}
	defer closeLog()

	if flags.pidFile != "" {
		if err := writeDaemonPIDFiles(flags.pidFile, flags.socket); err != nil {
			return daemonWrapf(ErrDaemonIO, "start.write-pid", err, "write pid file")
		}
		defer removeDaemonPIDFiles(flags.pidFile)
	}

	logSink.Printf("daemon starting id=%s profile=%s socket=%s pid=%d", settings.DaemonID, settings.Profile, flags.socket, os.Getpid())
	return serveAgentDaemon(ctx, flags, settings, logSink)
}

// runDaemonStartBackground forks the same binary in foreground mode and
// waits for the child's socket to become reachable before returning.
// This is the "self-spawn wrapper" pattern from M-2:
//
//   - PID file: written by the child in foreground mode (carries child PID).
//   - Log file: child writes to it directly. Parent does NOT open the log
//     file; if it did, both parent and child writing to the same file
//     would race and confuse log readers.
//   - Socket readiness: parent polls `net.Dial` on the socket; only
//     returns success once a connection is accepted.
//   - Child death before readiness: parent surfaces the wait error.
//   - Deadline: 15s. After that the parent kills the child and errors out.
//
// We intentionally do NOT double-fork. macOS launchd / systemd / install
// scripts prefer to manage foreground processes themselves; this wrapper
// is for ad-hoc CLI invocation only.
func runDaemonStartBackground(ctx lifecycle.Context, flags startFlags) error {
	if err := ensureDaemonStartLockAvailable(ctx, flags.lockFile); err != nil {
		return err
	}

	childArgs := []string{"daemon", "start", "--foreground", "--socket", flags.socket}
	if flags.pidFile != "" {
		childArgs = append(childArgs, "--pid-file", flags.pidFile)
	}
	if flags.logFile != "" {
		childArgs = append(childArgs, "--log-file", flags.logFile)
	}
	if flags.lockFile != "" {
		childArgs = append(childArgs, "--lock-file", flags.lockFile)
	}

	cmd, err := daemonSpawnHelper(childArgs)
	if err != nil {
		return err
	}

	// Detach stdio. We MUST point child stdout/stderr at a real OS file
	// (here /dev/null) rather than `io.Discard` — `io.Discard` would
	// cause exec.Cmd to spawn a parent-resident copy goroutine, and
	// when the parent CLI process exits the pipe's read end closes,
	// delivering SIGPIPE to the child on its next log write and
	// killing the daemon. The same /dev/null fd is used for stdin so
	// the daemon never sees an interactive terminal.
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return daemonWrapf(ErrDaemonIO, "background.open-dev-null", err, "open /dev/null")
	}
	defer devNull.Close()
	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	setDaemonChildSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return daemonWrapf(ErrDaemonProcess, "background.spawn", err, "spawn daemon child")
	}

	// Wait for the child to bind its socket OR die OR time out.
	exitCh := make(chan error, 1)
	go func() { exitCh <- cmd.Wait() }()

	deadline := time.NewTimer(15 * time.Second)
	defer deadline.Stop()
	poll := time.NewTicker(50 * time.Millisecond)
	defer poll.Stop()

	for {
		select {
		case err := <-exitCh:
			return daemonWrapf(ErrDaemonProcess, "background.wait-ready", err, "daemon child exited before socket was ready")
		case <-deadline.C:
			_ = cmd.Process.Kill()
			return daemonErrorf(ErrDaemonSocket, "background.wait-ready", "daemon socket %s did not become ready within 15s", flags.socket)
		case <-poll.C:
			conn, err := net.DialTimeout("unix", flags.socket, 200*time.Millisecond)
			if err != nil {
				continue
			}
			_ = conn.Close()
			if err := ensureDaemonSocketOwnedByChild(flags, cmd.Process.Pid); err != nil {
				_ = cmd.Process.Kill()
				return err
			}
			return nil
		}
	}
}

func ensureDaemonStartLockAvailable(ctx lifecycle.Context, lockFile string) error {
	probeCtx, cancel := context.WithTimeout(ctx.Context(), 100*time.Millisecond)
	defer cancel()

	lock, err := c9lock.AcquireFile(probeCtx, lockFile)
	if err != nil {
		return daemonWrapf(ErrDaemonLock, "background.preflight-lock", err, "daemon already running or starting; singleton lock %s is held", lockFile)
	}
	if err := lock.Release(); err != nil {
		return daemonWrapf(ErrDaemonLock, "background.preflight-lock", err, "release daemon singleton probe lock %s", lockFile)
	}
	return nil
}

func ensureDaemonSocketOwnedByChild(flags startFlags, childPID int) error {
	if flags.pidFile == "" {
		return nil
	}
	raw, err := os.ReadFile(flags.pidFile)
	if err != nil {
		return daemonWrapf(ErrDaemonIO, "background.verify-child-pid", err, "read daemon pid file")
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		return daemonWrapf(ErrDaemonIO, "background.verify-child-pid", err, "parse daemon pid file")
	}
	if pid != childPID {
		return daemonErrorf(
			ErrDaemonLock,
			"background.verify-child-pid",
			"daemon socket %s is already served by pid %d; spawned child pid %d",
			flags.socket,
			pid,
			childPID,
		)
	}
	return nil
}

// openLogSink returns a Logger port for structured log lines. When
// logFile is empty, logs go to stderr. When set, they go to BOTH stderr
// and the file so test runners and operators can both observe.
func openLogSink(logFile string) (logging.Logger, func(), error) {
	if logFile == "" {
		return logging.NewWriterLogger(os.Stderr), func() {}, nil
	}
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}
	w := io.MultiWriter(os.Stderr, f)
	return logging.NewWriterLogger(w), func() { _ = f.Close() }, nil
}

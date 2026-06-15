package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/taskdbplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/supervisor"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolpolicy"
	c9lock "github.com/teamswyg/riido-daemon/internal/lock"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
	"github.com/teamswyg/riido-daemon/internal/workdir"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// daemonSpawnHelper builds the exec.Cmd that the background-mode
// wrapper uses to launch the daemon child. Production sets it via
// init() to invoke os.Executable() with the supplied args; tests
// override it (in TestMain) so the test binary can fork itself as the
// daemon child.
//
// Mutable, but only mutated either by package init() or by TestMain
// before any test goroutine starts. No mutex needed.
var daemonSpawnHelper = defaultDaemonSpawnHelper

func defaultDaemonSpawnHelper(args []string) (*exec.Cmd, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, daemonWrapf(ErrDaemonProcess, "spawn.locate-executable", err, "locate daemon binary")
	}
	return exec.Command(exe, args...), nil
}

// DaemonStatusSchemaVersion identifies the JSON shape returned by
// `riido daemon status` and `riido daemon health`.
const DaemonStatusSchemaVersion = "riido-agent-daemon-status.v1"

// daemonStatus is the JSON payload exposed by the agent daemon over its
// local Unix socket.
type daemonStatus struct {
	SchemaVersion  string                `json:"schema_version"`
	DaemonID       string                `json:"daemon_id"`
	DaemonVersion  string                `json:"daemon_version"`
	PID            int                   `json:"pid"`
	UptimeSeconds  int                   `json:"uptime_seconds"`
	Health         string                `json:"health"`
	Ready          bool                  `json:"ready"`
	Readiness      string                `json:"readiness"`
	Profile        string                `json:"profile"`
	ServerURL      string                `json:"server_url,omitempty"`
	DeviceName     string                `json:"device_name"`
	WorkspaceCount int                   `json:"workspace_count"`
	SocketPath     string                `json:"socket_path"`
	LogFile        string                `json:"log_file,omitempty"`
	PIDFile        string                `json:"pid_file,omitempty"`
	RunningTasks   int                   `json:"running_tasks"`
	Metrics        daemonMetrics         `json:"metrics"`
	Runtimes       []runtimeactor.Status `json:"runtimes"`
	StartedAt      string                `json:"started_at"`
}

type daemonMetrics struct {
	RuntimeCount        int `json:"runtime_count"`
	RuntimeResponding   int `json:"runtime_responding"`
	ProviderAvailable   int `json:"provider_available"`
	ProviderUnavailable int `json:"provider_unavailable"`
	RunningTasks        int `json:"running_tasks"`
}

func runDaemon(args []string) error {
	return runDaemonWithLifecycle(lifecycle.Background(), args)
}

// runDaemonWithContext lets tests cancel the foreground daemon by
// canceling ctx. It is a stdlib-compatibility wrapper around the daemon's
// named lifecycle context.
func runDaemonWithContext(ctx context.Context, args []string) error {
	return runDaemonWithLifecycle(lifecycle.FromContext(ctx), args)
}

func runDaemonWithLifecycle(ctx lifecycle.Context, args []string) error {
	if len(args) < 1 {
		printUsage()
		return daemonErrorf(ErrDaemonUsage, "run", "missing daemon subcommand")
	}
	switch args[0] {
	case "start":
		return runDaemonStart(ctx, args[1:])
	case "status":
		return runDaemonStatus(args[1:])
	case "health":
		return runDaemonHealth(args[1:])
	case "ready":
		return runDaemonReady(args[1:])
	case "metrics":
		return runDaemonMetrics(args[1:])
	case "stop":
		return runDaemonStop(args[1:])
	case "logs":
		return runDaemonLogs(args[1:])
	default:
		printUsage()
		return daemonErrorf(ErrDaemonUsage, "run", "unknown daemon subcommand: %s", args[0])
	}
}

// ---- start ----

type startFlags struct {
	foreground bool
	socket     string
	pidFile    string
	logFile    string
	lockFile   string
}

func parseStartFlags(args []string) (startFlags, error) {
	out := startFlags{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--foreground":
			out.foreground = true
		case "--socket":
			i++
			if i >= len(args) {
				return out, daemonErrorf(ErrDaemonUsage, "start.parse-flags", "--socket requires a path")
			}
			out.socket = args[i]
		case "--pid-file":
			i++
			if i >= len(args) {
				return out, daemonErrorf(ErrDaemonUsage, "start.parse-flags", "--pid-file requires a path")
			}
			out.pidFile = args[i]
		case "--log-file":
			i++
			if i >= len(args) {
				return out, daemonErrorf(ErrDaemonUsage, "start.parse-flags", "--log-file requires a path")
			}
			out.logFile = args[i]
		case "--lock-file":
			i++
			if i >= len(args) {
				return out, daemonErrorf(ErrDaemonUsage, "start.parse-flags", "--lock-file requires a path")
			}
			out.lockFile = args[i]
		case "--help", "-h":
			printUsage()
			return out, nil
		default:
			return out, daemonErrorf(ErrDaemonUsage, "start.parse-flags", "unknown argument: %s", args[i])
		}
	}
	return out, nil
}

func runDaemonStart(ctx lifecycle.Context, args []string) error {
	flags, err := parseStartFlags(args)
	if err != nil {
		return err
	}
	if flags.socket == "" {
		def, err := defaultAgentDaemonSocket()
		if err != nil {
			return err
		}
		flags.socket = def
	}
	if flags.foreground {
		return runDaemonStartForeground(ctx, flags)
	}
	return runDaemonStartBackground(ctx, flags)
}

// runDaemonStartForeground is the in-process daemon — it spawns the
// RuntimeActor, opens the socket, and serves until ctx is cancelled or
// SIGTERM/SIGINT/shutdown-request fires. The background wrapper
// re-invokes the same binary with --foreground to land in this path.
func runDaemonStartForeground(ctx lifecycle.Context, flags startFlags) error {
	settings, err := loadDaemonSettings()
	if err != nil {
		return err
	}
	if flags.lockFile == "" {
		lockPath, err := defaultDaemonLockPath()
		if err != nil {
			return err
		}
		flags.lockFile = lockPath
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
		if err := os.WriteFile(flags.pidFile, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
			return daemonWrapf(ErrDaemonIO, "start.write-pid", err, "write pid file")
		}
		defer func() { _ = os.Remove(flags.pidFile) }()
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
func runDaemonStartBackground(_ lifecycle.Context, flags startFlags) error {
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
			return nil
		}
	}
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

// serveAgentDaemon opens the Unix socket, constructs one RuntimeActor per
// built-in provider adapter, answers status/health, and returns when ctx is
// canceled or SIGTERM arrives.
func serveAgentDaemon(ctx lifecycle.Context, flags startFlags, settings daemonSettings, log logging.Logger) error {
	// Remove a stale socket from a previous run.
	_ = os.Remove(flags.socket)

	startedAt := time.Now()

	rtActors, err := newDaemonRuntimeActors(settings, builtinDaemonAdapters())
	if err != nil {
		return err
	}
	startedRuntimes := make([]*runtimeactor.Actor, 0, len(rtActors))
	for _, rt := range rtActors {
		if err := rt.Start(ctx.Context()); err != nil {
			shutdownCtx, cancel := lifecycle.DetachedShutdown(lifecycle.ShutdownForced, 5*time.Second)
			stopRuntimeActors(shutdownCtx, startedRuntimes, log)
			cancel()
			return daemonWrapf(ErrDaemonRuntime, "serve.start-runtime", err, "runtimeactor.Start")
		}
		startedRuntimes = append(startedRuntimes, rt)
	}
	defer func() {
		shutdownCtx, cancel := lifecycle.DetachedShutdown(lifecycle.ShutdownGraceful, 5*time.Second)
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
		shutdownCtx, cancel := lifecycle.DetachedShutdown(lifecycle.ShutdownGraceful, 5*time.Second)
		defer cancel()
		if err := supActor.Stop(shutdownCtx.Context()); err != nil {
			log.Printf("supervisor stop error: %v", err)
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
	signalCtx, stop := lifecycle.Notify(ctx.WithShutdownLevel(lifecycle.ShutdownGraceful), daemonInterruptSignals()...)
	defer stop()

	shutdownCh := make(chan lifecycle.ShutdownLevel, 1)
	done := make(chan struct{})
	go func() {
		select {
		case <-signalCtx.Done():
			log.Printf("daemon shutdown requested level=%s source=signal", signalCtx.ShutdownLevel())
		case level := <-shutdownCh:
			log.Printf("daemon shutdown requested level=%s source=socket", level)
		}
		_ = ln.Close()
		close(done)
	}()

	log.Printf("daemon listening on %s", flags.socket)
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-done:
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

func daemonRuntimeModels(provider string) []runtimeactor.RuntimeModel {
	switch strings.TrimSpace(provider) {
	case "codex":
		return codexRuntimeModels(os.UserHomeDir)
	default:
		return nil
	}
}

func codexRuntimeModels(userHome func() (string, error)) []runtimeactor.RuntimeModel {
	modelID := codexConfiguredModelID(userHome)
	if modelID == "" {
		return nil
	}
	return []runtimeactor.RuntimeModel{{
		ModelID:   modelID,
		Label:     modelID,
		IsDefault: true,
	}}
}

func codexConfiguredModelID(userHome func() (string, error)) string {
	if userHome == nil {
		return ""
	}
	home, err := userHome()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	body, err := os.ReadFile(filepath.Join(home, ".codex", "config.toml"))
	if err != nil {
		return ""
	}
	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, rawValue, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != "model" {
			continue
		}
		value := strings.TrimSpace(rawValue)
		if unquoted, err := strconv.Unquote(value); err == nil {
			return strings.TrimSpace(unquoted)
		}
		if commentAt := strings.Index(value, "#"); commentAt >= 0 {
			value = strings.TrimSpace(value[:commentAt])
			if unquoted, err := strconv.Unquote(value); err == nil {
				return strings.TrimSpace(unquoted)
			}
		}
		return strings.TrimSpace(value)
	}
	return ""
}

func daemonToolAutoApprover(settings daemonSettings) agentbridge.AutoApprover {
	return toolpolicy.PolicyAutoApprover(settings.PolicyBundleDoc, policy.TrustTierHost)
}

func daemonToolStartGate(settings daemonSettings) agentbridge.ToolStartGate {
	return toolpolicy.PolicyToolStartGate(settings.PolicyBundleDoc, policy.TrustTierHost)
}

func stopRuntimeActors(ctx lifecycle.Context, runtimes []*runtimeactor.Actor, log logging.Logger) {
	for _, rt := range runtimes {
		if err := rt.Stop(ctx.Context()); err != nil {
			log.Printf("runtimeactor stop error level=%s: %v", ctx.ShutdownLevel(), err)
		}
	}
}

func providerRuntimeID(daemonID, provider string) string {
	if provider == "" {
		return daemonID
	}
	return daemonID + ":" + provider
}

func buildDaemonControlPlane(settings daemonSettings, startedAt time.Time) (controlplane.TaskSourcePort, controlplane.TaskReporterPort, string, error) {
	if settings.SaaSURL != "" {
		plane, err := saasplane.New(saasplane.Config{
			BaseURL:      settings.SaaSURL,
			DaemonID:     settings.DaemonID,
			DeviceID:     settings.DeviceID,
			DeviceSecret: settings.DeviceSecret,
			Profile:      settings.Profile,
			AppVersion:   settings.DaemonVersion,
			PID:          os.Getpid(),
			StartedAt:    startedAt.UTC(),
		})
		if err != nil {
			return nil, nil, "", daemonWrapf(ErrDaemonControlPlane, "control-plane.saas", err, "controlplane: saas source")
		}
		return plane, plane, "saas", nil
	}
	if settings.TaskDBSourcePath != "" {
		if settings.TaskQueueDir != "" {
			return nil, nil, "", daemonErrorf(ErrDaemonConfig, "control-plane.validate-config", "%s cannot be combined with %s", envTaskDBSourcePath, envTaskQueueDir)
		}
		if settings.TaskReportDir != "" {
			return nil, nil, "", daemonErrorf(ErrDaemonConfig, "control-plane.validate-config", "%s cannot be combined with %s", envTaskDBSourcePath, envTaskReportDir)
		}
		plane, err := taskdbplane.New(settings.TaskDBSourcePath)
		if err != nil {
			return nil, nil, "", daemonWrapf(ErrDaemonControlPlane, "control-plane.taskdb", err, "controlplane: task DB source")
		}
		return plane, plane, "taskdb", nil
	}
	if settings.TaskQueueDir == "" {
		if settings.TaskReportDir != "" {
			return nil, nil, "", daemonErrorf(ErrDaemonConfig, "control-plane.validate-config", "%s requires %s", envTaskReportDir, envTaskQueueDir)
		}
		return controlplane.NewMemorySource(), controlplane.NewMemoryReporter(), "memory", nil
	}
	source, err := controlplane.NewFileQueueSource(settings.TaskQueueDir)
	if err != nil {
		return nil, nil, "", daemonWrapf(ErrDaemonControlPlane, "control-plane.file-source", err, "controlplane: file queue source")
	}
	reportDir := settings.TaskReportDir
	if reportDir == "" {
		reportDir = filepath.Join(settings.TaskQueueDir, "reports")
	}
	reporter, err := controlplane.NewFileReporter(reportDir)
	if err != nil {
		return nil, nil, "", daemonWrapf(ErrDaemonControlPlane, "control-plane.file-reporter", err, "controlplane: file reporter")
	}
	return source, reporter, "file", nil
}

func startWorkdirCleanupLoop(ctx lifecycle.Context, cleaner workdir.Cleaner, settings daemonSettings, log logging.Logger) func() {
	if settings.WorkdirRetention <= 0 {
		return func() {}
	}
	if settings.WorkdirCleanupEvery <= 0 {
		settings.WorkdirCleanupEvery = time.Hour
	}
	cleanupCtx, cancel := lifecycle.WithCancel(ctx)
	runCleanup := func() {
		cutoff := time.Now().UTC().Add(-settings.WorkdirRetention)
		result, err := cleaner.CleanupArchivedBefore(cleanupCtx.Context(), workdir.CleanupRequest{ArchivedBefore: cutoff})
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("workdir cleanup error: %v", err)
			}
			return
		}
		if len(result.Removed) > 0 {
			log.Printf("workdir cleanup removed=%d scanned=%d retention=%s", len(result.Removed), result.ScannedArchiveRecords, settings.WorkdirRetention)
		}
	}
	runCleanup()
	go func() {
		ticker := time.NewTicker(settings.WorkdirCleanupEvery)
		defer ticker.Stop()
		for {
			select {
			case <-cleanupCtx.Done():
				return
			case <-ticker.C:
				runCleanup()
			}
		}
	}()
	return cancel
}

// builtinDaemonAdapters returns the four shipped agent adapters. The
// daemon's RuntimeActor takes ownership of their Detect lifecycle.
func builtinDaemonAdapters() []agentbridge.Adapter {
	return []agentbridge.Adapter{
		bridgeClaudeAdapter{},
		bridgeCodexAdapter{},
		bridgeOpenClawAdapter{},
		bridgeCursorAdapter{},
	}
}

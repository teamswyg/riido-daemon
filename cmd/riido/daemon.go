package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
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
	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	c9lock "github.com/teamswyg/riido-daemon/internal/lock"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/process/childreg"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
	"github.com/teamswyg/riido-daemon/internal/workdir"
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
		return nil, fmt.Errorf("locate daemon binary: %w", err)
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
	return runDaemonWithContext(context.Background(), args)
}

// runDaemonWithContext lets tests cancel the foreground daemon by
// canceling ctx. Production callers pass context.Background().
func runDaemonWithContext(ctx context.Context, args []string) error {
	if len(args) < 1 {
		printUsage()
		return fmt.Errorf("missing daemon subcommand")
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
		return fmt.Errorf("unknown daemon subcommand: %s", args[0])
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
				return out, errors.New("--socket requires a path")
			}
			out.socket = args[i]
		case "--pid-file":
			i++
			if i >= len(args) {
				return out, errors.New("--pid-file requires a path")
			}
			out.pidFile = args[i]
		case "--log-file":
			i++
			if i >= len(args) {
				return out, errors.New("--log-file requires a path")
			}
			out.logFile = args[i]
		case "--lock-file":
			i++
			if i >= len(args) {
				return out, errors.New("--lock-file requires a path")
			}
			out.lockFile = args[i]
		case "--help", "-h":
			printUsage()
			return out, nil
		default:
			return out, fmt.Errorf("unknown argument: %s", args[i])
		}
	}
	return out, nil
}

func runDaemonStart(ctx context.Context, args []string) error {
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
func runDaemonStartForeground(ctx context.Context, flags startFlags) error {
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
	if flags.pidFile == "" {
		pidPath, err := defaultDaemonPidPath()
		if err != nil {
			return err
		}
		flags.pidFile = pidPath
	}
	lock, alreadyRunning, err := acquireDaemonSingleton(flags.lockFile, flags.pidFile)
	if err != nil {
		return fmt.Errorf("acquire daemon singleton lock %s: %w", flags.lockFile, err)
	}
	if alreadyRunning {
		fmt.Fprintln(os.Stderr, "riido daemon: another instance is already running; exiting")
		return nil
	}
	defer lock.Release()

	logSink, closeLog, err := openLogSink(flags.logFile)
	if err != nil {
		return fmt.Errorf("open log sink: %w", err)
	}
	defer closeLog()

	// Record our PID so a future start can probe liveness for stale-lock reclaim.
	if err := os.WriteFile(flags.pidFile, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		return fmt.Errorf("write pid file: %w", err)
	}
	defer func() { _ = os.Remove(flags.pidFile) }()

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
func runDaemonStartBackground(_ context.Context, flags startFlags) error {
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
		return fmt.Errorf("open /dev/null: %w", err)
	}
	defer devNull.Close()
	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	setDaemonChildSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("spawn daemon child: %w", err)
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
			return fmt.Errorf("daemon child exited before socket was ready: %v", err)
		case <-deadline.C:
			_ = cmd.Process.Kill()
			return fmt.Errorf("daemon socket %s did not become ready within 15s", flags.socket)
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
func serveAgentDaemon(ctx context.Context, flags startFlags, settings daemonSettings, log logging.Logger) error {
	// Anti-hijack (D4): never unlink a socket a live daemon is still serving.
	// The singleton lock already fast-fails a same-identity duplicate; this guards
	// the transition window where a differently-pathed daemon shares the socket.
	if daemonSocketServing(flags.socket) {
		log.Printf("daemon: another instance is already serving %s; stepping aside", flags.socket)
		return nil
	}
	// Remove a stale socket from a previous run.
	_ = os.Remove(flags.socket)

	startedAt := time.Now()

	// D6: reap provider process groups orphaned by a previous unclean daemon
	// exit (SIGKILL/crash), then track newly spawned children so they too can be
	// reaped on the next start. Only the singleton lock holder reaches here.
	regPath := childRegistryPath(flags)
	if reaped, rerr := childreg.ReapOrphans(regPath); rerr != nil {
		log.Printf("orphan provider reaper: %v", rerr)
	} else if reaped > 0 {
		log.Printf("reaped %d orphan provider process group(s) from a previous run", reaped)
	}
	childReg := childreg.New(regPath)

	rtActors, err := newDaemonRuntimeActors(settings, builtinDaemonAdapters(), childReg)
	if err != nil {
		return err
	}
	startedRuntimes := make([]*runtimeactor.Actor, 0, len(rtActors))
	for _, rt := range rtActors {
		if err := rt.Start(ctx); err != nil {
			stopRuntimeActors(context.Background(), startedRuntimes, log)
			return fmt.Errorf("runtimeactor.Start: %w", err)
		}
		startedRuntimes = append(startedRuntimes, rt)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
		TextFlushBytes:      settings.TextFlushBytes,
		TextFlushInterval:   settings.TextFlushInterval,
		PolicyBundleVersion: settings.PolicyBundle,
		PolicyBundle:        settings.PolicyBundleDoc,
		RuntimeTrustTier:    policy.TrustTierHost,
	})
	if err != nil {
		return fmt.Errorf("supervisor.New: %w", err)
	}
	if err := supActor.Start(ctx); err != nil {
		return fmt.Errorf("supervisor.Start: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := supActor.Stop(shutdownCtx); err != nil {
			log.Printf("supervisor stop error: %v", err)
		}
	}()
	log.Printf("supervisor started workdir_root=%s control_plane=%s queue_dir=%s report_dir=%s", settings.WorkdirRoot, controlPlaneKind, settings.TaskQueueDir, settings.TaskReportDir)
	stopCleanup := startWorkdirCleanupLoop(ctx, workdirAdapter, settings, log)
	defer stopCleanup()

	ln, err := net.Listen("unix", flags.socket)
	if err != nil {
		return fmt.Errorf("listen %s: %w", flags.socket, err)
	}
	defer func() {
		_ = ln.Close()
		_ = os.Remove(flags.socket)
	}()

	// Honor ctx cancellation, POSIX signals, AND in-socket "shutdown"
	// requests together. The shutdown channel is buffered so the
	// handler's non-blocking send can never deadlock if multiple
	// clients race a stop request.
	signalCtx, stop := signal.NotifyContext(ctx, daemonInterruptSignals()...)
	defer stop()

	shutdownCh := make(chan struct{}, 1)
	done := make(chan struct{})
	go func() {
		select {
		case <-signalCtx.Done():
		case <-shutdownCh:
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

func newDaemonRuntimeActors(settings daemonSettings, adapters []agentbridge.Adapter, childReg *childreg.Registry) ([]*runtimeactor.Actor, error) {
	out := make([]*runtimeactor.Actor, 0, len(adapters))
	for _, adapter := range adapters {
		name := strings.TrimSpace(adapter.Name())
		if name == "" {
			return nil, errors.New("runtimeactor.New: adapter name is required")
		}
		rt, err := newDaemonRuntimeActor(settings, providerRuntimeID(settings.DaemonID, name), adapter, settings.RuntimeAgents, childReg)
		if err != nil {
			return nil, fmt.Errorf("runtimeactor.New(%s): %w", name, err)
		}
		out = append(out, rt)
	}
	if len(out) == 0 {
		return nil, errors.New("runtimeactor.New: at least one adapter is required")
	}
	return out, nil
}

func newDaemonRuntimeActor(settings daemonSettings, runtimeID string, adapter agentbridge.Adapter, agents []runtimeactor.AgentStatus, childReg *childreg.Registry) (*runtimeactor.Actor, error) {
	return runtimeactor.New(runtimeactor.Config{
		RuntimeID:           runtimeID,
		Owner:               settings.RuntimeOwner,
		DeviceName:          settings.DeviceName,
		Agents:              agents,
		Models:              daemonRuntimeModels(adapter.Name()),
		Adapters:            []agentbridge.Adapter{adapter},
		Process:             processexec.NewWithObserver(childReg),
		MaxConcurrent:       1,
		HardTimeout:         settings.RunHardTimeout,
		SemanticIdle:        settings.RunSemanticIdle,
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

func stopRuntimeActors(ctx context.Context, runtimes []*runtimeactor.Actor, log logging.Logger) {
	for _, rt := range runtimes {
		if err := rt.Stop(ctx); err != nil {
			log.Printf("runtimeactor stop error: %v", err)
		}
	}
}

func providerRuntimeID(daemonID string, provider string) string {
	if provider == "" {
		return daemonID
	}
	return daemonID + ":" + provider
}

func buildDaemonControlPlane(settings daemonSettings, startedAt time.Time) (controlplane.TaskSourcePort, controlplane.TaskReporterPort, string, error) {
	if settings.SaaSURL != "" {
		plane, err := saasplane.New(saasplane.Config{
			BaseURL:         settings.SaaSURL,
			DaemonID:        settings.DaemonID,
			DeviceID:        settings.DeviceID,
			DeviceSecret:    settings.DeviceSecret,
			Profile:         settings.Profile,
			AppVersion:      settings.DaemonVersion,
			PID:             os.Getpid(),
			StartedAt:       startedAt.UTC(),
			ClaimWaitMs:     settings.ClaimWaitMs,
			LongPollTimeout: settings.LongPollTimeout,
		})
		if err != nil {
			return nil, nil, "", fmt.Errorf("controlplane: saas source: %w", err)
		}
		return plane, plane, "saas", nil
	}
	if settings.TaskDBSourcePath != "" {
		if settings.TaskQueueDir != "" {
			return nil, nil, "", fmt.Errorf("%s cannot be combined with %s", envTaskDBSourcePath, envTaskQueueDir)
		}
		if settings.TaskReportDir != "" {
			return nil, nil, "", fmt.Errorf("%s cannot be combined with %s", envTaskDBSourcePath, envTaskReportDir)
		}
		plane, err := taskdbplane.New(settings.TaskDBSourcePath)
		if err != nil {
			return nil, nil, "", fmt.Errorf("controlplane: task DB source: %w", err)
		}
		return plane, plane, "taskdb", nil
	}
	if settings.TaskQueueDir == "" {
		if settings.TaskReportDir != "" {
			return nil, nil, "", fmt.Errorf("%s requires %s", envTaskReportDir, envTaskQueueDir)
		}
		return controlplane.NewMemorySource(), controlplane.NewMemoryReporter(), "memory", nil
	}
	source, err := controlplane.NewFileQueueSource(settings.TaskQueueDir)
	if err != nil {
		return nil, nil, "", fmt.Errorf("controlplane: file queue source: %w", err)
	}
	reportDir := settings.TaskReportDir
	if reportDir == "" {
		reportDir = filepath.Join(settings.TaskQueueDir, "reports")
	}
	reporter, err := controlplane.NewFileReporter(reportDir)
	if err != nil {
		return nil, nil, "", fmt.Errorf("controlplane: file reporter: %w", err)
	}
	return source, reporter, "file", nil
}

func startWorkdirCleanupLoop(ctx context.Context, cleaner workdir.Cleaner, settings daemonSettings, log logging.Logger) func() {
	if settings.WorkdirRetention <= 0 {
		return func() {}
	}
	if settings.WorkdirCleanupEvery <= 0 {
		settings.WorkdirCleanupEvery = time.Hour
	}
	cleanupCtx, cancel := context.WithCancel(ctx)
	runCleanup := func() {
		cutoff := time.Now().UTC().Add(-settings.WorkdirRetention)
		result, err := cleaner.CleanupArchivedBefore(cleanupCtx, workdir.CleanupRequest{ArchivedBefore: cutoff})
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

// daemonRequest is the JSON envelope read off the socket.
type daemonRequest struct {
	Method string `json:"method"`
}

func handleDaemonConn(conn net.Conn, flags startFlags, settings daemonSettings, startedAt time.Time, runtimes []*runtimeactor.Actor, shutdownCh chan<- struct{}, log logging.Logger) {
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
		case shutdownCh <- struct{}{}:
		default:
		}
		log.Printf("shutdown request received")
	default:
		_ = json.NewEncoder(conn).Encode(map[string]any{"error": "unknown method", "method": req.Method})
	}
}

func writeShutdownAck(conn net.Conn) {
	_ = json.NewEncoder(conn).Encode(map[string]string{
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
	_ = json.NewEncoder(conn).Encode(s)
}

func writeHealth(conn net.Conn) {
	_ = json.NewEncoder(conn).Encode(map[string]string{
		"schema_version": DaemonStatusSchemaVersion,
		"health":         "ok",
	})
}

func writeReady(conn net.Conn, runtimes []*runtimeactor.Actor) {
	obs := observeDaemon(runtimes)
	_ = json.NewEncoder(conn).Encode(map[string]any{
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
	_ = json.NewEncoder(conn).Encode(map[string]any{
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	obs := daemonObservation{
		runtimes: make([]runtimeactor.Status, 0, len(runtimes)),
		metrics:  daemonMetrics{RuntimeCount: len(runtimes)},
	}
	for _, rt := range runtimes {
		rtStatus, err := rt.Status(ctx)
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

// ---- status / health (client side) ----

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
				return "", errors.New("--socket requires a path")
			}
			return args[i], nil
		}
	}
	return defaultAgentDaemonSocket()
}

// defaultDaemonAppDataRoot is the single identity directory the daemon derives
// its socket, lock, and pid file from, so they always refer to the same daemon
// instance (no cross-path socket hijack — D4).
func defaultDaemonAppDataRoot() (hostintegration.AppDataRoot, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return hostintegration.AppDataRoot{}, err
	}
	return hostintegration.DefaultAppDataRoot(hostintegration.AppDataRootInput{
		Channel:  hostintegration.DistributionChannelDevLocal,
		HostOS:   hostintegration.HostOSDarwin,
		UserHome: home,
	})
}

func defaultAgentDaemonSocket() (string, error) {
	root, err := defaultDaemonAppDataRoot()
	if err != nil {
		return "", err
	}
	endpoint, err := hostintegration.DefaultLocalIPCEndpoint(hostintegration.LocalIPCEndpointInput{
		Channel:     hostintegration.DistributionChannelDevLocal,
		HostOS:      hostintegration.HostOSDarwin,
		AppDataRoot: root,
		Owner:       hostintegration.LocalIPCOwnerHelper,
		Name:        "agentd",
	})
	if err != nil {
		return "", err
	}
	return endpoint.Path, nil
}

// defaultDaemonLockPath co-locates the singleton lock with the socket in the
// app-data identity root so a manual `riido daemon start` and a desktop-launched
// daemon resolve to the SAME single-instance lock (machine = one daemon).
func defaultDaemonLockPath() (string, error) {
	root, err := defaultDaemonAppDataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root.Path, "daemon.lock"), nil
}

func defaultDaemonPidPath() (string, error) {
	root, err := defaultDaemonAppDataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root.Path, "daemon.pid"), nil
}

// childRegistryPath co-locates the orphan-reaper registry (D6) with the daemon's
// pid file, so it lands in the same identity dir (desktop userData or app-data)
// and a crash leaves it where the next start looks.
func childRegistryPath(flags startFlags) string {
	base := strings.TrimSpace(flags.pidFile)
	if base == "" {
		base = strings.TrimSpace(flags.lockFile)
	}
	if base == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(base), "daemon-children.pids")
}

// acquireDaemonSingleton tries to become the single daemon instance (D3). It
// returns the held lock, or (nil, true, nil) when another LIVE daemon already
// owns the lock ("already running" — the caller exits cleanly rather than
// blocking as a lock-waiter zombie). A lock left by a provably-dead owner (e.g.
// a Windows ".claim" file after a crash) is reclaimed and retried once; on Unix
// flock auto-releases on death so that reclaim path is unreachable.
func acquireDaemonSingleton(lockPath, pidPath string) (*c9lock.FileLock, bool, error) {
	lock, err := c9lock.TryAcquireFile(lockPath)
	if err == nil {
		return lock, false, nil
	}
	if !errors.Is(err, c9lock.ErrLocked) {
		return nil, false, err
	}
	// Reclaim only when the recorded owner is provably dead; never reclaim a lock
	// whose liveness we cannot determine (conservative — avoids double daemons).
	if pid, ok := readDaemonPID(pidPath); ok && !daemonPIDProbablyAlive(pid) {
		_ = c9lock.RemoveStaleLock(lockPath)
		_ = os.Remove(pidPath)
		if lock, err = c9lock.TryAcquireFile(lockPath); err == nil {
			return lock, false, nil
		}
	}
	return nil, true, nil
}

func readDaemonPID(pidPath string) (int, bool) {
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

// daemonSocketServing reports whether a daemon is actively listening on the unix
// socket. A successful connect means a live peer owns it, so the socket must not
// be unlinked (doing so would orphan that daemon — D4).
func daemonSocketServing(socketPath string) bool {
	conn, err := net.DialTimeout("unix", socketPath, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func daemonCall(sock string, method string) error {
	conn, err := net.DialTimeout("unix", sock, 2*time.Second)
	if err != nil {
		return fmt.Errorf("dial %s: %w", sock, err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if err := json.NewEncoder(conn).Encode(daemonRequest{Method: method}); err != nil {
		return fmt.Errorf("encode request: %w", err)
	}
	body, err := io.ReadAll(conn)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("read response: %w", err)
	}
	_, err = os.Stdout.Write(body)
	return err
}

// ---- stop ----

func runDaemonStop(args []string) error {
	socket := ""
	pidFile := ""
	timeoutSeconds := 5
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--socket":
			i++
			if i >= len(args) {
				return errors.New("--socket requires a path")
			}
			socket = args[i]
		case "--pid-file":
			i++
			if i >= len(args) {
				return errors.New("--pid-file requires a path")
			}
			pidFile = args[i]
		case "--timeout-seconds":
			i++
			if i >= len(args) {
				return errors.New("--timeout-seconds requires a value")
			}
			v, err := strconv.Atoi(args[i])
			if err != nil || v <= 0 {
				return fmt.Errorf("--timeout-seconds must be positive int: %v", args[i])
			}
			timeoutSeconds = v
		case "--help", "-h":
			printUsage()
			return nil
		default:
			return fmt.Errorf("unknown argument: %s", args[i])
		}
	}
	if socket == "" && pidFile == "" {
		return errors.New("daemon stop requires at least one of --socket or --pid-file")
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
		return fmt.Errorf("daemon stop: socket %s did not respond and --pid-file is not provided", socket)
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
		return fmt.Errorf("read pid file: %w", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		return fmt.Errorf("parse pid: %w", err)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	if err := signalDaemonProcessTerm(proc); err != nil {
		return fmt.Errorf("terminate daemon process: %w", err)
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !daemonProcessExists(proc) {
			return nil // gone
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err := signalDaemonProcessKill(proc); err != nil {
		return fmt.Errorf("kill daemon process: %w", err)
	}
	return nil
}

// ---- logs ----

func runDaemonLogs(args []string) error {
	logFile := ""
	lines := 50
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--log-file":
			i++
			if i >= len(args) {
				return errors.New("--log-file requires a path")
			}
			logFile = args[i]
		case "--lines":
			i++
			if i >= len(args) {
				return errors.New("--lines requires a value")
			}
			v, err := strconv.Atoi(args[i])
			if err != nil || v <= 0 {
				return fmt.Errorf("--lines must be positive int")
			}
			lines = v
		case "--help", "-h":
			printUsage()
			return nil
		default:
			return fmt.Errorf("unknown argument: %s", args[i])
		}
	}
	if logFile == "" {
		return errors.New("--log-file is required")
	}
	f, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
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
		return err
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

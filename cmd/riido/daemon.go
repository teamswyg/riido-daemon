package main

import (
	"context"
	"os"
	"os/exec"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
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
	switch daemonCommand(args[0]) {
	case daemonCommandStart:
		return runDaemonStart(ctx, args[1:])
	case daemonCommandStatus:
		return runDaemonStatus(args[1:])
	case daemonCommandHealth:
		return runDaemonHealth(args[1:])
	case daemonCommandReady:
		return runDaemonReady(args[1:])
	case daemonCommandMetrics:
		return runDaemonMetrics(args[1:])
	case daemonCommandStop:
		return runDaemonStop(args[1:])
	case daemonCommandLogs:
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
	if flags.lockFile == "" {
		lockPath, err := defaultDaemonLockPath()
		if err != nil {
			return err
		}
		flags.lockFile = lockPath
	}
	if flags.foreground {
		return runDaemonStartForeground(ctx, flags)
	}
	return runDaemonStartBackground(ctx, flags)
}

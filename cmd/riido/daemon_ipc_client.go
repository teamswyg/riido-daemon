package main

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func runDaemonStatus(args []string) error {
	return runDaemonSocketCommand(args, daemonMethodStatus)
}

func runDaemonHealth(args []string) error {
	return runDaemonSocketCommand(args, daemonMethodHealth)
}

func runDaemonReady(args []string) error {
	return runDaemonSocketCommand(args, daemonMethodReady)
}

func runDaemonMetrics(args []string) error {
	return runDaemonSocketCommand(args, daemonMethodMetrics)
}

func runDaemonSocketCommand(args []string, method daemonMethod) error {
	sock, err := requireSocketFlag(args)
	if isCLIHelp(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return daemonCall(sock, method)
}

func requireSocketFlag(args []string) (string, error) {
	if len(args) == 0 {
		return defaultAgentDaemonSocket()
	}
	if isHelpArg(args[0]) {
		printUsage()
		return "", errCLIHelp
	}
	if args[0] != "--socket" {
		return "", daemonErrorf(ErrDaemonUsage, "ipc.parse-socket", "unknown argument: %s", args[0])
	}
	if len(args) < 2 {
		return "", daemonErrorf(ErrDaemonUsage, "ipc.parse-socket", "--socket requires a path")
	}
	if len(args) > 2 {
		return "", daemonErrorf(ErrDaemonUsage, "ipc.parse-socket", "unknown argument: %s", args[2])
	}
	return args[1], nil
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

func daemonCall(sock string, method daemonMethod) error {
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
	force := false
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
		case "--force":
			force = true
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
	level := lifecycle.ShutdownGraceful
	if force {
		level = lifecycle.ShutdownForced
	}

	// 1. Socket shutdown first (preferred — cooperative, no signals).
	if socket != "" {
		if ok := tryShutdownViaSocket(socket, timeout, level); ok {
			return nil
		}
	}

	// 2. PID SIGTERM fallback.
	if pidFile == "" {
		return daemonErrorf(ErrDaemonSocket, "stop.socket-fallback", "daemon stop: socket %s did not respond and --pid-file is not provided", socket)
	}
	return stopViaPIDFile(pidFile, timeout, level)
}

package main

import (
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
)

func defaultAgentDaemonSocket() (string, error) {
	return defaultAgentDaemonSocketForOS(currentDaemonHostOS(), os.UserHomeDir, os.UserCacheDir)
}

func defaultAgentDaemonSocketForOS(goos string, userHome, userCache daemonPathRootResolver) (string, error) {
	input, err := defaultDaemonAppDataRootInput(goos, userHome, userCache)
	if err != nil {
		return "", daemonWrapf(ErrDaemonIO, "ipc.default-socket.host-root", err, "resolve default host root")
	}
	root, err := hostintegration.DefaultAppDataRoot(input)
	if err != nil {
		return "", daemonWrapf(ErrDaemonConfig, "ipc.default-socket.app-data-root", err, "resolve default app data root")
	}
	endpoint, err := hostintegration.DefaultLocalIPCEndpoint(hostintegration.LocalIPCEndpointInput{
		Channel:     hostintegration.DistributionChannelDevLocal,
		HostOS:      root.HostOS,
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

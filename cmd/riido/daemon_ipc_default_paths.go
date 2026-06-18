package main

import (
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
)

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

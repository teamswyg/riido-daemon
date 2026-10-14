package main

import (
	"os"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
)

func defaultAgentDaemonWorkdirRoot(userHome func() (string, error)) (string, error) {
	return defaultAgentDaemonWorkdirRootForOS(currentDaemonHostOS(), userHome, os.UserCacheDir)
}

func defaultAgentDaemonWorkdirRootForOS(goos string, userHome, userCache daemonPathRootResolver) (string, error) {
	input, err := defaultDaemonAppDataRootInput(goos, userHome, userCache)
	if err != nil {
		return "", daemonWrapf(ErrDaemonIO, "settings.default-workdir.host-root", err, "resolve default host root")
	}
	root, err := hostintegration.DefaultAppDataRoot(input)
	if err != nil {
		return "", daemonWrapf(ErrDaemonConfig, "settings.default-workdir.app-data-root", err, "resolve default app data root")
	}
	return root.WorkdirRoot(), nil
}

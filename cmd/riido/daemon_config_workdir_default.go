package main

import "github.com/teamswyg/riido-daemon/internal/hostintegration"

func defaultAgentDaemonWorkdirRoot(userHome func() (string, error)) (string, error) {
	home, err := userHome()
	if err != nil {
		return "", daemonWrapf(ErrDaemonIO, "settings.default-workdir.user-home", err, "resolve user home")
	}
	root, err := hostintegration.DefaultAppDataRoot(hostintegration.AppDataRootInput{
		Channel:  hostintegration.DistributionChannelDevLocal,
		HostOS:   hostintegration.HostOSDarwin,
		UserHome: home,
	})
	if err != nil {
		return "", daemonWrapf(ErrDaemonConfig, "settings.default-workdir.app-data-root", err, "resolve default app data root")
	}
	return root.WorkdirRoot(), nil
}

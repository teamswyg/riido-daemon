package main

import (
	"fmt"
	"runtime"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
)

type daemonPathRootResolver func() (string, error)

func defaultDaemonAppDataRootInput(
	goos string,
	userHome daemonPathRootResolver,
	userCache daemonPathRootResolver,
) (hostintegration.AppDataRootInput, error) {
	hostOS := hostintegration.HostOSDarwin
	input := hostintegration.AppDataRootInput{
		Channel: hostintegration.DistributionChannelDevLocal,
		HostOS:  hostOS,
	}
	if goos == "windows" {
		hostOS = hostintegration.HostOSWindows
		localAppData, err := userCache()
		if err != nil {
			return hostintegration.AppDataRootInput{}, fmt.Errorf("resolve Windows local app data: %w", err)
		}
		input.HostOS = hostOS
		input.WindowsLocalAppDataRoot = localAppData
	} else {
		home, err := userHome()
		if err != nil {
			return hostintegration.AppDataRootInput{}, fmt.Errorf("resolve user home: %w", err)
		}
		input.UserHome = home
	}
	return input, nil
}

func currentDaemonHostOS() string {
	return runtime.GOOS
}

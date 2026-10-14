package hostintegration

import (
	"errors"
	"strings"
)

func windowsLocalAppDataRoot(in AppDataRootInput) (AppDataRoot, error) {
	root := strings.TrimSpace(in.WindowsLocalAppDataRoot)
	if root == "" {
		return AppDataRoot{}, errors.New("windows dev-local app data root requires WindowsLocalAppDataRoot")
	}
	return AppDataRoot{
		Channel: in.Channel,
		HostOS:  in.HostOS,
		Scope:   AppDataRootWindowsLocalAppData,
		Path:    joinHostPath(in.HostOS, root, "Riido"),
	}, nil
}

func windowsPackageLocalRoot(in AppDataRootInput) (AppDataRoot, error) {
	root := strings.TrimSpace(in.WindowsPackageLocalDataRoot)
	if root == "" {
		return AppDataRoot{}, errors.New("msix app data root requires WindowsPackageLocalDataRoot")
	}
	return AppDataRoot{
		Channel: in.Channel,
		HostOS:  in.HostOS,
		Scope:   AppDataRootWindowsPackageLocal,
		Path:    root,
	}, nil
}

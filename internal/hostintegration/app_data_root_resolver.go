package hostintegration

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// DefaultAppDataRoot resolves the default app data root for a distribution
// channel. Store-managed channels require adapter-provided container/package
// roots and never derive from arbitrary user home scanning.
func DefaultAppDataRoot(in AppDataRootInput) (AppDataRoot, error) {
	if !in.Channel.Valid() {
		return AppDataRoot{}, fmt.Errorf("unknown distribution channel %q", in.Channel)
	}
	if !in.HostOS.Valid() {
		return AppDataRoot{}, fmt.Errorf("unknown host OS %q", in.HostOS)
	}

	switch in.Channel {
	case DistributionChannelDevLocal:
		return devLocalAppDataRoot(in)
	case DistributionChannelDeveloperID:
		if in.HostOS != HostOSDarwin {
			return AppDataRoot{}, errors.New("developer-id app data root requires darwin host OS")
		}
		return darwinApplicationSupportRoot(in)
	case DistributionChannelMacAppStore:
		if in.HostOS != HostOSDarwin {
			return AppDataRoot{}, errors.New("mac-app-store app data root requires darwin host OS")
		}
		return darwinStoreAppDataRoot(in)
	case DistributionChannelMSIXSideload, DistributionChannelMSIXStore:
		if in.HostOS != HostOSWindows {
			return AppDataRoot{}, errors.New("msix app data root requires windows host OS")
		}
		return windowsPackageLocalRoot(in)
	default:
		return AppDataRoot{}, fmt.Errorf("unsupported distribution channel %q", in.Channel)
	}
}

func devLocalAppDataRoot(in AppDataRootInput) (AppDataRoot, error) {
	switch in.HostOS {
	case HostOSDarwin:
		return darwinApplicationSupportRoot(in)
	case HostOSWindows:
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
	default:
		return AppDataRoot{}, fmt.Errorf("unsupported dev-local host OS %q", in.HostOS)
	}
}

func darwinApplicationSupportRoot(in AppDataRootInput) (AppDataRoot, error) {
	home := strings.TrimSpace(in.UserHome)
	if home == "" {
		return AppDataRoot{}, errors.New("darwin app data root requires UserHome")
	}
	return AppDataRoot{
		Channel: in.Channel,
		HostOS:  in.HostOS,
		Scope:   AppDataRootUserApplicationSupport,
		Path:    filepath.Join(home, "Library", "Application Support", "riido"),
	}, nil
}

func darwinStoreAppDataRoot(in AppDataRootInput) (AppDataRoot, error) {
	if root := strings.TrimSpace(in.DarwinAppGroupRoot); root != "" {
		return AppDataRoot{
			Channel: in.Channel,
			HostOS:  in.HostOS,
			Scope:   AppDataRootAppGroup,
			Path:    root,
		}, nil
	}
	if root := strings.TrimSpace(in.DarwinSandboxContainerRoot); root != "" {
		return AppDataRoot{
			Channel: in.Channel,
			HostOS:  in.HostOS,
			Scope:   AppDataRootSandboxContainer,
			Path:    root,
		}, nil
	}
	return AppDataRoot{}, errors.New("mac-app-store app data root requires app group or sandbox container root")
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

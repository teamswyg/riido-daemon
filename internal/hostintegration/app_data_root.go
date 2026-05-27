package hostintegration

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// HostOS is the operating-system family that owns app data path semantics.
type HostOS string

const (
	HostOSDarwin  HostOS = "darwin"
	HostOSWindows HostOS = "windows"
)

// AppDataRootScope records why a root path is acceptable for a distribution
// channel. Store channels must not silently fall back to unmanaged home paths.
type AppDataRootScope string

const (
	AppDataRootUserApplicationSupport AppDataRootScope = "user-application-support"
	AppDataRootSandboxContainer       AppDataRootScope = "sandbox-container"
	AppDataRootAppGroup               AppDataRootScope = "app-group"
	AppDataRootWindowsLocalAppData    AppDataRootScope = "windows-local-app-data"
	AppDataRootWindowsPackageLocal    AppDataRootScope = "windows-package-local-data"
)

// AppDataRootInput is supplied by an OS adapter. C11 validates the selected
// root; it does not inspect OS entitlements or call platform APIs itself.
type AppDataRootInput struct {
	Channel DistributionChannel
	HostOS  HostOS

	UserHome                    string
	DarwinSandboxContainerRoot  string
	DarwinAppGroupRoot          string
	WindowsLocalAppDataRoot     string
	WindowsPackageLocalDataRoot string
}

// AppDataRoot is the channel-approved local data root for Riido control-plane
// state on a customer machine.
type AppDataRoot struct {
	Channel DistributionChannel
	HostOS  HostOS
	Scope   AppDataRootScope
	Path    string
}

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

// WorkdirRoot returns the C6 workdir root under the approved app data root.
func (r AppDataRoot) WorkdirRoot() string {
	return joinHostPath(r.HostOS, r.Path, "workspaces")
}

// Valid reports whether os is one of the SSOT-defined host OS values.
func (os HostOS) Valid() bool {
	switch os {
	case HostOSDarwin, HostOSWindows:
		return true
	default:
		return false
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

func joinHostPath(os HostOS, elems ...string) string {
	if os != HostOSWindows {
		return filepath.Join(elems...)
	}
	return joinWindowsPath(elems...)
}

func joinWindowsPath(elems ...string) string {
	var parts []string
	for i, elem := range elems {
		trimmed := strings.TrimSpace(elem)
		if trimmed == "" {
			continue
		}
		if i == 0 {
			parts = append(parts, strings.TrimRight(trimmed, `\/`))
			continue
		}
		parts = append(parts, strings.Trim(trimmed, `\/`))
	}
	return strings.Join(parts, `\`)
}

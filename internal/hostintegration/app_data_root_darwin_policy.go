package hostintegration

import (
	"errors"
	"path/filepath"
	"strings"
)

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
		return darwinAppGroupRoot(in, root), nil
	}
	if root := strings.TrimSpace(in.DarwinSandboxContainerRoot); root != "" {
		return darwinSandboxContainerRoot(in, root), nil
	}
	return AppDataRoot{}, errors.New("mac-app-store app data root requires app group or sandbox container root")
}

func darwinAppGroupRoot(in AppDataRootInput, root string) AppDataRoot {
	return AppDataRoot{
		Channel: in.Channel,
		HostOS:  in.HostOS,
		Scope:   AppDataRootAppGroup,
		Path:    root,
	}
}

func darwinSandboxContainerRoot(in AppDataRootInput, root string) AppDataRoot {
	return AppDataRoot{
		Channel: in.Channel,
		HostOS:  in.HostOS,
		Scope:   AppDataRootSandboxContainer,
		Path:    root,
	}
}

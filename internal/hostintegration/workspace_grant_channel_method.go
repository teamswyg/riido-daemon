package hostintegration

import (
	"errors"
	"fmt"
)

func validateWorkspaceGrantChannelMethod(channel DistributionChannel, hostOS HostOS, method WorkspaceGrantMethod) error {
	switch channel {
	case DistributionChannelDevLocal:
		return validateDevLocalWorkspaceGrantMethod(channel, method)
	case DistributionChannelDeveloperID:
		return validateDeveloperIDWorkspaceGrantMethod(channel, hostOS, method)
	case DistributionChannelMacAppStore:
		return validateMacAppStoreWorkspaceGrantMethod(hostOS, method)
	case DistributionChannelMSIXSideload:
		return validateMSIXSideloadWorkspaceGrantMethod(hostOS, method)
	case DistributionChannelMSIXStore:
		return validateMSIXStoreWorkspaceGrantMethod(hostOS, method)
	}
	return nil
}

func validateDevLocalWorkspaceGrantMethod(channel DistributionChannel, method WorkspaceGrantMethod) error {
	if method == WorkspaceGrantDevLocalPath || method == WorkspaceGrantUserSelectedFolder {
		return nil
	}
	return fmt.Errorf("%s workspace grant requires dev-local path or user-selected folder", channel)
}

func validateDeveloperIDWorkspaceGrantMethod(channel DistributionChannel, hostOS HostOS, method WorkspaceGrantMethod) error {
	if hostOS != HostOSDarwin {
		return errors.New("developer-id workspace grant requires darwin host OS")
	}
	if method == WorkspaceGrantUserSelectedFolder || method == WorkspaceGrantSecurityScopedBookmark {
		return nil
	}
	return fmt.Errorf("%s workspace grant requires user-selected folder or security-scoped bookmark", channel)
}

func validateMacAppStoreWorkspaceGrantMethod(hostOS HostOS, method WorkspaceGrantMethod) error {
	if hostOS != HostOSDarwin {
		return errors.New("mac-app-store workspace grant requires darwin host OS")
	}
	if method != WorkspaceGrantSecurityScopedBookmark {
		return errors.New("mac-app-store workspace grant requires security-scoped bookmark")
	}
	return nil
}

func validateMSIXSideloadWorkspaceGrantMethod(hostOS HostOS, method WorkspaceGrantMethod) error {
	if hostOS != HostOSWindows {
		return errors.New("msix-sideload workspace grant requires windows host OS")
	}
	if method != WorkspaceGrantUserSelectedFolder && method != WorkspaceGrantWindowsFolderPickerGrant {
		return errors.New("msix-sideload workspace grant requires user-selected folder or windows folder picker grant")
	}
	return nil
}

func validateMSIXStoreWorkspaceGrantMethod(hostOS HostOS, method WorkspaceGrantMethod) error {
	if hostOS != HostOSWindows {
		return errors.New("msix-store workspace grant requires windows host OS")
	}
	if method != WorkspaceGrantWindowsFolderPickerGrant {
		return errors.New("msix-store workspace grant requires windows folder picker grant")
	}
	return nil
}

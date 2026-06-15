package hostintegration

import (
	"fmt"
)

func validateHelperRuntimeAppDataRoot(channel DistributionChannel, root AppDataRoot) error {
	if root.Channel != channel {
		return fmt.Errorf("helper app data root channel %q does not match runtime channel %q", root.Channel, channel)
	}
	switch channel {
	case DistributionChannelDeveloperID, DistributionChannelDevLocal:
	case DistributionChannelMacAppStore:
		if root.Scope != AppDataRootAppGroup && root.Scope != AppDataRootSandboxContainer {
			return fmt.Errorf("mac-app-store helper requires app-group or sandbox-container app data root, got %q", root.Scope)
		}
	case DistributionChannelMSIXSideload, DistributionChannelMSIXStore:
		if root.Scope != AppDataRootWindowsPackageLocal {
			return fmt.Errorf("msix helper requires windows-package-local-data app data root, got %q", root.Scope)
		}
	}
	return nil
}

// Valid reports whether role is one of the SSOT-defined helper roles.
func (role HelperRuntimeRole) Valid() bool {
	switch role {
	case HelperRuntimeRoleLocalBroker,
		HelperRuntimeRoleSandboxedLoginItem,
		HelperRuntimeRoleMSIXPackagedBroker,
		HelperRuntimeRoleMSIXFullTrustTray:
		return true
	default:
		return false
	}
}

// Valid reports whether rule is one of the SSOT-defined background rules.
func (rule HelperBackgroundRule) Valid() bool {
	switch rule {
	case HelperBackgroundExplicitConsent,
		HelperBackgroundConsentAndStoreReview:
		return true
	default:
		return false
	}
}

// Valid reports whether registration is one of the SSOT-defined helper startup
// registration families.
func (registration HelperStartupRegistration) Valid() bool {
	switch registration {
	case HelperStartupLaunchAgentOrLoginItem,
		HelperStartupServiceManagementLoginItem,
		HelperStartupMSIXPackagedStartupTask:
		return true
	default:
		return false
	}
}

package policy

import "github.com/teamswyg/riido-daemon/internal/hostintegration"

func evaluateWindowsServiceInstall(input StoreChannelPolicyInput) Decision {
	switch input.Channel {
	case hostintegration.DistributionChannelMSIXSideload:
		return deny("STORE_CHANNEL_SURFACE_DISCOURAGED", "Windows service install is avoided for msix-sideload by default")
	case hostintegration.DistributionChannelMSIXStore:
		return deny("STORE_CHANNEL_SURFACE_FORBIDDEN", "Windows service install is forbidden by default for msix-store")
	default:
		return deny("STORE_CHANNEL_SURFACE_NOT_APPLICABLE", "Windows service install is not a macOS distribution surface")
	}
}

package policy

import "github.com/teamswyg/riido-daemon/internal/hostintegration"

func evaluateSelfUpdater(input StoreChannelPolicyInput) Decision {
	switch input.Channel {
	case hostintegration.DistributionChannelDeveloperID,
		hostintegration.DistributionChannelDevLocal,
		hostintegration.DistributionChannelMSIXSideload:
		return allowStoreChannelSurface("self-updater is allowed for this non-store-managed distribution channel")
	case hostintegration.DistributionChannelMacAppStore,
		hostintegration.DistributionChannelMSIXStore:
		return deny("STORE_CHANNEL_SURFACE_FORBIDDEN", "store-managed channels must use store update mechanisms")
	default:
		return deny("STORE_CHANNEL_UNKNOWN_CHANNEL", "store channel policy requires a known distribution channel")
	}
}

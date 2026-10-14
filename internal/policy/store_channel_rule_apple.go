package policy

import "github.com/teamswyg/riido-daemon/internal/hostintegration"

func evaluateProviderCLIUserSelectedPath(input StoreChannelPolicyInput) Decision {
	if input.Channel == hostintegration.DistributionChannelMacAppStore && !input.OSGrantPresent {
		return deny("STORE_CHANNEL_REQUIRES_OS_GRANT", "mac-app-store provider CLI path requires a sandbox, user-selected executable, or security-scoped grant")
	}
	if input.Channel == hostintegration.DistributionChannelMacAppStore && !input.StoreReviewApproved {
		return deny("STORE_CHANNEL_REQUIRES_STORE_REVIEW_APPROVAL", macAppStoreProviderCLIReviewReason+" and review notes")
	}
	return allowStoreChannelSurface("user-selected provider CLI path is allowed for this distribution channel")
}

func evaluateDirectLaunchAgentInstall(input StoreChannelPolicyInput) Decision {
	switch input.Channel {
	case hostintegration.DistributionChannelDeveloperID, hostintegration.DistributionChannelDevLocal:
		if !input.ExplicitConsentGranted {
			return deny("STORE_CHANNEL_REQUIRES_CONSENT", "direct LaunchAgent install requires explicit user consent")
		}
		return allowStoreChannelSurface("direct LaunchAgent install is allowed for this macOS distribution channel")
	case hostintegration.DistributionChannelMacAppStore:
		return deny("STORE_CHANNEL_SURFACE_FORBIDDEN", "direct LaunchAgent install is forbidden for mac-app-store")
	default:
		return deny("STORE_CHANNEL_SURFACE_NOT_APPLICABLE", "direct LaunchAgent install is not a Windows distribution surface")
	}
}

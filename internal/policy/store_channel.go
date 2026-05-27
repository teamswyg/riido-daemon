package policy

import "github.com/teamswyg/riido-daemon/internal/hostintegration"

// StoreSurface is one host/distribution surface covered by
// distribution-host-integration.md §6.
type StoreSurface string

const (
	StoreSurfaceProviderCLIBundling         StoreSurface = "provider-cli-bundling"
	StoreSurfaceProviderCLIUserSelectedPath StoreSurface = "provider-cli-user-selected-path"
	StoreSurfaceSilentProviderAutoInstall   StoreSurface = "silent-provider-auto-install"
	StoreSurfaceBackgroundHelper            StoreSurface = "background-helper"
	StoreSurfaceDirectLaunchAgentInstall    StoreSurface = "direct-launch-agent-install"
	StoreSurfaceWindowsServiceInstall       StoreSurface = "windows-service-install"
	StoreSurfaceExternalTCPListener         StoreSurface = "external-tcp-listener"
	StoreSurfaceLocalIPC                    StoreSurface = "local-ipc"
	StoreSurfaceSelfUpdater                 StoreSurface = "self-updater"
	StoreSurfaceArbitraryHomeScan           StoreSurface = "arbitrary-home-scan"
)

// StoreChannelPolicyInput asks whether a distribution channel may use one
// concrete host surface. OSGrantPresent means the caller has already reduced
// platform-specific proof such as a security-scoped bookmark or package-local
// grant into a boolean fact.
type StoreChannelPolicyInput struct {
	Channel                hostintegration.DistributionChannel
	Surface                StoreSurface
	ExplicitConsentGranted bool
	OSGrantPresent         bool
	StoreReviewApproved    bool
}

// EvaluateStoreChannelPolicy implements distribution-host-integration.md §6 as
// a pure C7 decision. It does not install helpers, open IPC endpoints, inspect
// entitlements, or execute provider CLIs.
func EvaluateStoreChannelPolicy(input StoreChannelPolicyInput) Decision {
	if !input.Channel.Valid() {
		return deny("STORE_CHANNEL_UNKNOWN_CHANNEL", "store channel policy requires a known distribution channel")
	}
	if !input.Surface.Valid() {
		return deny("STORE_CHANNEL_UNKNOWN_SURFACE", "store channel policy requires a known host surface")
	}

	switch input.Surface {
	case StoreSurfaceProviderCLIBundling,
		StoreSurfaceSilentProviderAutoInstall,
		StoreSurfaceExternalTCPListener,
		StoreSurfaceArbitraryHomeScan:
		return deny("STORE_CHANNEL_SURFACE_FORBIDDEN", "this host surface is forbidden for every distribution channel")
	case StoreSurfaceProviderCLIUserSelectedPath:
		return evaluateProviderCLIUserSelectedPath(input)
	case StoreSurfaceBackgroundHelper:
		return evaluateBackgroundHelper(input)
	case StoreSurfaceDirectLaunchAgentInstall:
		return evaluateDirectLaunchAgentInstall(input)
	case StoreSurfaceWindowsServiceInstall:
		return evaluateWindowsServiceInstall(input)
	case StoreSurfaceLocalIPC:
		return allowStoreChannelSurface("local IPC is allowed when the C11 adapter keeps it local-only")
	case StoreSurfaceSelfUpdater:
		return evaluateSelfUpdater(input)
	default:
		return deny("STORE_CHANNEL_UNKNOWN_SURFACE", "store channel policy requires a known host surface")
	}
}

// Valid reports whether surface is one of the SSOT-defined store channel
// policy rows.
func (surface StoreSurface) Valid() bool {
	switch surface {
	case StoreSurfaceProviderCLIBundling,
		StoreSurfaceProviderCLIUserSelectedPath,
		StoreSurfaceSilentProviderAutoInstall,
		StoreSurfaceBackgroundHelper,
		StoreSurfaceDirectLaunchAgentInstall,
		StoreSurfaceWindowsServiceInstall,
		StoreSurfaceExternalTCPListener,
		StoreSurfaceLocalIPC,
		StoreSurfaceSelfUpdater,
		StoreSurfaceArbitraryHomeScan:
		return true
	default:
		return false
	}
}

func evaluateProviderCLIUserSelectedPath(input StoreChannelPolicyInput) Decision {
	if input.Channel == hostintegration.DistributionChannelMacAppStore && !input.OSGrantPresent {
		return deny("STORE_CHANNEL_REQUIRES_OS_GRANT", "mac-app-store provider CLI path requires a sandbox, user-selected executable, or security-scoped grant")
	}
	if input.Channel == hostintegration.DistributionChannelMacAppStore && !input.StoreReviewApproved {
		return deny("STORE_CHANNEL_REQUIRES_STORE_REVIEW_APPROVAL", "mac-app-store provider CLI execution requires App Review approval and review notes")
	}
	return allowStoreChannelSurface("user-selected provider CLI path is allowed for this distribution channel")
}

func evaluateBackgroundHelper(input StoreChannelPolicyInput) Decision {
	if !input.ExplicitConsentGranted {
		return deny("STORE_CHANNEL_REQUIRES_CONSENT", "background helper requires explicit user consent")
	}
	if input.Channel.StoreManaged() && !input.StoreReviewApproved {
		return deny("STORE_CHANNEL_REQUIRES_STORE_REVIEW_APPROVAL", "store-managed background helper requires store policy review approval")
	}
	return allowStoreChannelSurface("background helper is allowed for this distribution channel")
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

func allowStoreChannelSurface(reason string) Decision {
	return Decision{Allowed: true, Code: "ALLOWED", Reason: reason}
}

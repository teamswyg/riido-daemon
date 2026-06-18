package policy

const macAppStoreProviderCLIReviewReason = "mac-app-store provider CLI execution requires App Review approval"

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
	}
	return deny("STORE_CHANNEL_UNKNOWN_SURFACE", "store channel policy requires a known host surface")
}

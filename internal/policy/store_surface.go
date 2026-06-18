package policy

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

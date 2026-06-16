package hostintegration

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

// WorkdirRoot returns the C6 workdir root under the approved app data root.
func (r AppDataRoot) WorkdirRoot() string {
	return joinHostPath(r.HostOS, r.Path, "workspaces")
}

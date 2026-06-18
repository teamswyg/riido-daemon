package hostintegration

// HelperRuntimePlan is the C11 runtime shape a Store App/helper adapter may
// implement. Provider CLI binaries are intentionally outside this plan.
type HelperRuntimePlan struct {
	Channel                       DistributionChannel
	HostOS                        HostOS
	Role                          HelperRuntimeRole
	Endpoint                      LocalIPCEndpoint
	AppDataRoot                   AppDataRoot
	BackgroundRule                HelperBackgroundRule
	StartupRegistration           HelperStartupRegistration
	BackgroundAllowed             bool
	RequiresStoreReviewApproval   bool
	RequiresWorkspaceGrant        WorkspaceGrantMethod
	ProviderCLIBundlingAllowed    bool
	DirectLaunchAgentAllowed      bool
	WindowsServiceAllowed         bool
	SharedLocationInstallAllowed  bool
	StandaloneCodeDownloadAllowed bool
	SelfUpdaterAllowed            bool
	StoreManagedUpdates           bool
	ReviewNoteSurfaces            []string
}

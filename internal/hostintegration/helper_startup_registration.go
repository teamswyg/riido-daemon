package hostintegration

// HelperStartupRegistration records the OS registration family a packaged
// helper adapter may use after user consent.
type HelperStartupRegistration string

const (
	HelperStartupLaunchAgentOrLoginItem     HelperStartupRegistration = "launch-agent-or-login-item"
	HelperStartupServiceManagementLoginItem HelperStartupRegistration = "service-management-login-item"
	HelperStartupMSIXPackagedStartupTask    HelperStartupRegistration = "msix-packaged-startup-task"
)

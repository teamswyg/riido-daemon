package hostintegration

// HelperRuntimeRole is the packaged role shape for the local helper/broker.
// C11 owns this host-facing shape; C4 owns provider process execution.
type HelperRuntimeRole string

const (
	HelperRuntimeRoleLocalBroker        HelperRuntimeRole = "local-helper-broker"
	HelperRuntimeRoleSandboxedLoginItem HelperRuntimeRole = "sandboxed-login-item-helper"
	HelperRuntimeRoleMSIXPackagedBroker HelperRuntimeRole = "msix-packaged-helper-broker"
	HelperRuntimeRoleMSIXFullTrustTray  HelperRuntimeRole = "msix-packaged-full-trust-helper-tray"
)

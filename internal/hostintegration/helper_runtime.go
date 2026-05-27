package hostintegration

import (
	"fmt"
	"strings"
)

// HelperRuntimeRole is the packaged role shape for the local helper/broker.
// C11 owns this host-facing shape; C4 owns provider process execution.
type HelperRuntimeRole string

const (
	HelperRuntimeRoleLocalBroker        HelperRuntimeRole = "local-helper-broker"
	HelperRuntimeRoleSandboxedLoginItem HelperRuntimeRole = "sandboxed-login-item-helper"
	HelperRuntimeRoleMSIXPackagedBroker HelperRuntimeRole = "msix-packaged-helper-broker"
	HelperRuntimeRoleMSIXFullTrustTray  HelperRuntimeRole = "msix-packaged-full-trust-helper-tray"
)

// HelperBackgroundRule records which external approval facts are required
// before the helper may run after the foreground app exits.
type HelperBackgroundRule string

const (
	HelperBackgroundExplicitConsent       HelperBackgroundRule = "explicit-consent"
	HelperBackgroundConsentAndStoreReview HelperBackgroundRule = "explicit-consent-and-store-review"
)

// HelperStartupRegistration records the OS registration family a packaged
// helper adapter may use after user consent.
type HelperStartupRegistration string

const (
	HelperStartupLaunchAgentOrLoginItem     HelperStartupRegistration = "launch-agent-or-login-item"
	HelperStartupServiceManagementLoginItem HelperStartupRegistration = "service-management-login-item"
	HelperStartupMSIXPackagedStartupTask    HelperStartupRegistration = "msix-packaged-startup-task"
)

// HelperRuntimePlanInput is reduced by C11 adapters before they install or
// start any helper process. It is pure data and does not call OS APIs.
type HelperRuntimePlanInput struct {
	Channel             DistributionChannel
	HostOS              HostOS
	AppDataRoot         AppDataRoot
	Consent             ConsentState
	StoreReviewApproved bool
	EndpointName        string
}

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

// ResolveHelperRuntimePlan returns the helper/tray role and local-only IPC
// contract for a distribution channel. It does not spawn providers, install
// startup entries, open sockets/pipes, or inspect platform entitlements.
func ResolveHelperRuntimePlan(in HelperRuntimePlanInput) (HelperRuntimePlan, error) {
	if !in.Channel.Valid() {
		return HelperRuntimePlan{}, fmt.Errorf("unknown distribution channel %q", in.Channel)
	}
	if !in.HostOS.Valid() {
		return HelperRuntimePlan{}, fmt.Errorf("unknown host OS %q", in.HostOS)
	}
	if err := validateHelperRuntimeAppDataRoot(in.Channel, in.AppDataRoot); err != nil {
		return HelperRuntimePlan{}, err
	}

	name := strings.TrimSpace(in.EndpointName)
	if name == "" {
		name = "riido"
	}
	endpoint, err := DefaultLocalIPCEndpoint(LocalIPCEndpointInput{
		Channel:     in.Channel,
		HostOS:      in.HostOS,
		AppDataRoot: in.AppDataRoot,
		Owner:       LocalIPCOwnerHelper,
		Name:        name,
	})
	if err != nil {
		return HelperRuntimePlan{}, err
	}

	plan := HelperRuntimePlan{
		Channel:                       in.Channel,
		HostOS:                        in.HostOS,
		Endpoint:                      endpoint,
		AppDataRoot:                   in.AppDataRoot,
		BackgroundRule:                HelperBackgroundExplicitConsent,
		ProviderCLIBundlingAllowed:    false,
		WindowsServiceAllowed:         false,
		SharedLocationInstallAllowed:  false,
		StandaloneCodeDownloadAllowed: false,
		SelfUpdaterAllowed:            true,
	}

	switch in.Channel {
	case DistributionChannelDevLocal, DistributionChannelDeveloperID:
		plan.Role = HelperRuntimeRoleLocalBroker
		plan.StartupRegistration = HelperStartupLaunchAgentOrLoginItem
		plan.BackgroundAllowed = in.Consent.BackgroundHelper
		plan.DirectLaunchAgentAllowed = true
	case DistributionChannelMacAppStore:
		plan.Role = HelperRuntimeRoleSandboxedLoginItem
		plan.BackgroundRule = HelperBackgroundConsentAndStoreReview
		plan.StartupRegistration = HelperStartupServiceManagementLoginItem
		plan.BackgroundAllowed = in.Consent.BackgroundHelper && in.StoreReviewApproved
		plan.RequiresStoreReviewApproval = true
		plan.RequiresWorkspaceGrant = WorkspaceGrantSecurityScopedBookmark
		plan.SelfUpdaterAllowed = false
		plan.StoreManagedUpdates = true
		plan.ReviewNoteSurfaces = []string{
			"app-sandbox-entitlement-review-notes",
			"service-management-login-item-consent",
			"security-scoped-workspace-grant",
			"helper-purpose-review-note",
			"provider-non-bundling-review-note",
			"review-demo-mode",
			"privacy-metadata-allowlist",
		}
	case DistributionChannelMSIXSideload:
		plan.Role = HelperRuntimeRoleMSIXPackagedBroker
		plan.StartupRegistration = HelperStartupMSIXPackagedStartupTask
		plan.BackgroundAllowed = in.Consent.BackgroundHelper
	case DistributionChannelMSIXStore:
		plan.Role = HelperRuntimeRoleMSIXFullTrustTray
		plan.BackgroundRule = HelperBackgroundConsentAndStoreReview
		plan.StartupRegistration = HelperStartupMSIXPackagedStartupTask
		plan.BackgroundAllowed = in.Consent.BackgroundHelper && in.StoreReviewApproved
		plan.RequiresStoreReviewApproval = true
		plan.SelfUpdaterAllowed = false
		plan.StoreManagedUpdates = true
		plan.ReviewNoteSurfaces = []string{
			"runfulltrust-review-note",
			"partner-center-review-notes",
			"provider-non-bundling-review-note",
			"review-demo-mode",
			"privacy-metadata-allowlist",
		}
	default:
		return HelperRuntimePlan{}, fmt.Errorf("unsupported distribution channel %q", in.Channel)
	}

	return plan, nil
}

func validateHelperRuntimeAppDataRoot(channel DistributionChannel, root AppDataRoot) error {
	if root.Channel != channel {
		return fmt.Errorf("helper app data root channel %q does not match runtime channel %q", root.Channel, channel)
	}
	switch channel {
	case DistributionChannelMacAppStore:
		if root.Scope != AppDataRootAppGroup && root.Scope != AppDataRootSandboxContainer {
			return fmt.Errorf("mac-app-store helper requires app-group or sandbox-container app data root, got %q", root.Scope)
		}
	case DistributionChannelMSIXSideload, DistributionChannelMSIXStore:
		if root.Scope != AppDataRootWindowsPackageLocal {
			return fmt.Errorf("msix helper requires windows-package-local-data app data root, got %q", root.Scope)
		}
	}
	return nil
}

// Valid reports whether role is one of the SSOT-defined helper roles.
func (role HelperRuntimeRole) Valid() bool {
	switch role {
	case HelperRuntimeRoleLocalBroker,
		HelperRuntimeRoleSandboxedLoginItem,
		HelperRuntimeRoleMSIXPackagedBroker,
		HelperRuntimeRoleMSIXFullTrustTray:
		return true
	default:
		return false
	}
}

// Valid reports whether rule is one of the SSOT-defined background rules.
func (rule HelperBackgroundRule) Valid() bool {
	switch rule {
	case HelperBackgroundExplicitConsent,
		HelperBackgroundConsentAndStoreReview:
		return true
	default:
		return false
	}
}

// Valid reports whether registration is one of the SSOT-defined helper startup
// registration families.
func (registration HelperStartupRegistration) Valid() bool {
	switch registration {
	case HelperStartupLaunchAgentOrLoginItem,
		HelperStartupServiceManagementLoginItem,
		HelperStartupMSIXPackagedStartupTask:
		return true
	default:
		return false
	}
}

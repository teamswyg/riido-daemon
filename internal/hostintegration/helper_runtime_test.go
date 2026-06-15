package hostintegration

import (
	"slices"
	"testing"
)

func TestResolveHelperRuntimePlanMacAppStoreUsesSandboxedLoginItem(t *testing.T) {
	root := mustDarwinStoreAppDataRoot(t, DistributionChannelMacAppStore)

	plan, err := ResolveHelperRuntimePlan(HelperRuntimePlanInput{
		Channel:             DistributionChannelMacAppStore,
		HostOS:              HostOSDarwin,
		AppDataRoot:         root,
		Consent:             ConsentState{BackgroundHelper: true},
		StoreReviewApproved: true,
		EndpointName:        "agentd",
	})
	if err != nil {
		t.Fatalf("ResolveHelperRuntimePlan: %v", err)
	}

	if plan.Role != HelperRuntimeRoleSandboxedLoginItem {
		t.Fatalf("role = %q, want %q", plan.Role, HelperRuntimeRoleSandboxedLoginItem)
	}
	if plan.StartupRegistration != HelperStartupServiceManagementLoginItem {
		t.Fatalf("startup registration = %q", plan.StartupRegistration)
	}
	if plan.Endpoint.EndpointKind != LocalIPCEndpointUnixSocket {
		t.Fatalf("endpoint kind = %q", plan.Endpoint.EndpointKind)
	}
	if plan.Endpoint.Owner != LocalIPCOwnerHelper {
		t.Fatalf("endpoint owner = %q", plan.Endpoint.Owner)
	}
	if plan.Endpoint.Path != "/Users/tester/Library/Group Containers/group.io.riido/agentd.sock" {
		t.Fatalf("endpoint path = %q", plan.Endpoint.Path)
	}
	if plan.RequiresWorkspaceGrant != WorkspaceGrantSecurityScopedBookmark {
		t.Fatalf("workspace grant = %q", plan.RequiresWorkspaceGrant)
	}
	if !plan.BackgroundAllowed {
		t.Fatal("mac-app-store background should be allowed with consent and review approval")
	}
	if !plan.RequiresStoreReviewApproval {
		t.Fatal("mac-app-store login item must require App Store review approval")
	}
	if plan.ProviderCLIBundlingAllowed {
		t.Fatal("mac-app-store helper plan must not allow bundled provider CLIs")
	}
	if plan.DirectLaunchAgentAllowed {
		t.Fatal("mac-app-store helper plan must not allow direct LaunchAgent install")
	}
	if plan.SharedLocationInstallAllowed {
		t.Fatal("mac-app-store helper plan must not allow shared-location code install")
	}
	if plan.StandaloneCodeDownloadAllowed {
		t.Fatal("mac-app-store helper plan must not allow standalone code download")
	}
	if plan.SelfUpdaterAllowed {
		t.Fatal("mac-app-store helper plan must use App Store-managed updates")
	}
	if !hasReviewSurface(plan.ReviewNoteSurfaces, "helper-purpose-review-note") {
		t.Fatalf("review surfaces missing helper-purpose-review-note: %v", plan.ReviewNoteSurfaces)
	}
	if !hasReviewSurface(plan.ReviewNoteSurfaces, "service-management-login-item-consent") {
		t.Fatalf("review surfaces missing service-management-login-item-consent: %v", plan.ReviewNoteSurfaces)
	}
}

func TestResolveHelperRuntimePlanMacAppStoreBackgroundRequiresConsentAndReview(t *testing.T) {
	root := mustDarwinStoreAppDataRoot(t, DistributionChannelMacAppStore)
	tests := []struct {
		name        string
		consent     bool
		review      bool
		wantAllowed bool
	}{
		{name: "no consent no review"},
		{name: "consent without review", consent: true},
		{name: "review without consent", review: true},
		{name: "consent and review", consent: true, review: true, wantAllowed: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := ResolveHelperRuntimePlan(HelperRuntimePlanInput{
				Channel:             DistributionChannelMacAppStore,
				HostOS:              HostOSDarwin,
				AppDataRoot:         root,
				Consent:             ConsentState{BackgroundHelper: tt.consent},
				StoreReviewApproved: tt.review,
			})
			if err != nil {
				t.Fatalf("ResolveHelperRuntimePlan: %v", err)
			}
			if plan.BackgroundRule != HelperBackgroundConsentAndStoreReview {
				t.Fatalf("background rule = %q", plan.BackgroundRule)
			}
			if plan.BackgroundAllowed != tt.wantAllowed {
				t.Fatalf("background allowed = %v, want %v", plan.BackgroundAllowed, tt.wantAllowed)
			}
		})
	}
}

func TestResolveHelperRuntimePlanRejectsMacAppStoreHomeFallback(t *testing.T) {
	_, err := ResolveHelperRuntimePlan(HelperRuntimePlanInput{
		Channel: DistributionChannelMacAppStore,
		HostOS:  HostOSDarwin,
		AppDataRoot: AppDataRoot{
			Channel: DistributionChannelMacAppStore,
			HostOS:  HostOSDarwin,
			Scope:   AppDataRootUserApplicationSupport,
			Path:    "/Users/tester/Library/Application Support/riido",
		},
	})
	if err == nil {
		t.Fatal("expected mac-app-store helper plan to reject unmanaged home fallback")
	}
}

func TestResolveHelperRuntimePlanMSIXStoreUsesFullTrustTrayNamedPipe(t *testing.T) {
	root := mustMSIXAppDataRoot(t, DistributionChannelMSIXStore)

	plan, err := ResolveHelperRuntimePlan(HelperRuntimePlanInput{
		Channel:             DistributionChannelMSIXStore,
		HostOS:              HostOSWindows,
		AppDataRoot:         root,
		Consent:             ConsentState{BackgroundHelper: true},
		StoreReviewApproved: true,
		EndpointName:        "agentd",
	})
	if err != nil {
		t.Fatalf("ResolveHelperRuntimePlan: %v", err)
	}

	if plan.Role != HelperRuntimeRoleMSIXFullTrustTray {
		t.Fatalf("role = %q, want %q", plan.Role, HelperRuntimeRoleMSIXFullTrustTray)
	}
	if plan.Endpoint.EndpointKind != LocalIPCEndpointNamedPipe {
		t.Fatalf("endpoint kind = %q", plan.Endpoint.EndpointKind)
	}
	if plan.Endpoint.Path != `\\.\pipe\riido-msix-store-helper-agentd` {
		t.Fatalf("endpoint path = %q", plan.Endpoint.Path)
	}
	if plan.AppDataRoot.Scope != AppDataRootWindowsPackageLocal {
		t.Fatalf("app data scope = %q", plan.AppDataRoot.Scope)
	}
	if !plan.BackgroundAllowed {
		t.Fatal("msix-store background should be allowed with consent and review approval")
	}
	if !plan.RequiresStoreReviewApproval {
		t.Fatal("msix-store full-trust helper must require Store review approval")
	}
	if plan.ProviderCLIBundlingAllowed {
		t.Fatal("msix-store helper plan must not allow bundled provider CLIs")
	}
	if plan.WindowsServiceAllowed {
		t.Fatal("msix-store helper plan must not allow Windows service install by default")
	}
	if plan.SelfUpdaterAllowed {
		t.Fatal("msix-store helper plan must use Store-managed updates")
	}
	if !hasReviewSurface(plan.ReviewNoteSurfaces, "runfulltrust-review-note") {
		t.Fatalf("review surfaces missing runfulltrust-review-note: %v", plan.ReviewNoteSurfaces)
	}
}

func TestResolveHelperRuntimePlanMSIXStoreBackgroundRequiresConsentAndReview(t *testing.T) {
	root := mustMSIXAppDataRoot(t, DistributionChannelMSIXStore)
	tests := []struct {
		name        string
		consent     bool
		review      bool
		wantAllowed bool
	}{
		{name: "no consent no review"},
		{name: "consent without review", consent: true},
		{name: "review without consent", review: true},
		{name: "consent and review", consent: true, review: true, wantAllowed: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := ResolveHelperRuntimePlan(HelperRuntimePlanInput{
				Channel:             DistributionChannelMSIXStore,
				HostOS:              HostOSWindows,
				AppDataRoot:         root,
				Consent:             ConsentState{BackgroundHelper: tt.consent},
				StoreReviewApproved: tt.review,
			})
			if err != nil {
				t.Fatalf("ResolveHelperRuntimePlan: %v", err)
			}
			if plan.BackgroundRule != HelperBackgroundConsentAndStoreReview {
				t.Fatalf("background rule = %q", plan.BackgroundRule)
			}
			if plan.BackgroundAllowed != tt.wantAllowed {
				t.Fatalf("background allowed = %v, want %v", plan.BackgroundAllowed, tt.wantAllowed)
			}
		})
	}
}

func TestResolveHelperRuntimePlanMSIXSideloadUsesPackagedBrokerWithoutStoreReview(t *testing.T) {
	root := mustMSIXAppDataRoot(t, DistributionChannelMSIXSideload)

	plan, err := ResolveHelperRuntimePlan(HelperRuntimePlanInput{
		Channel:      DistributionChannelMSIXSideload,
		HostOS:       HostOSWindows,
		AppDataRoot:  root,
		Consent:      ConsentState{BackgroundHelper: true},
		EndpointName: "agentd",
	})
	if err != nil {
		t.Fatalf("ResolveHelperRuntimePlan: %v", err)
	}

	if plan.Role != HelperRuntimeRoleMSIXPackagedBroker {
		t.Fatalf("role = %q, want %q", plan.Role, HelperRuntimeRoleMSIXPackagedBroker)
	}
	if plan.BackgroundRule != HelperBackgroundExplicitConsent {
		t.Fatalf("background rule = %q", plan.BackgroundRule)
	}
	if !plan.BackgroundAllowed {
		t.Fatal("msix-sideload background should only require explicit consent")
	}
	if plan.RequiresStoreReviewApproval {
		t.Fatal("msix-sideload helper must not require Store review approval")
	}
	if plan.WindowsServiceAllowed {
		t.Fatal("msix-sideload helper plan must not allow Windows service install by default")
	}
	if !plan.SelfUpdaterAllowed {
		t.Fatal("msix-sideload may use a non-Store update mechanism")
	}
}

func TestResolveHelperRuntimePlanRejectsMSIXHomeFallback(t *testing.T) {
	_, err := ResolveHelperRuntimePlan(HelperRuntimePlanInput{
		Channel: DistributionChannelMSIXStore,
		HostOS:  HostOSWindows,
		AppDataRoot: AppDataRoot{
			Channel: DistributionChannelMSIXStore,
			HostOS:  HostOSWindows,
			Scope:   AppDataRootWindowsLocalAppData,
			Path:    `C:\Users\tester\AppData\Local\Riido`,
		},
	})
	if err == nil {
		t.Fatal("expected msix-store helper plan to reject non-package app data root")
	}
}

func mustMSIXAppDataRoot(t *testing.T, channel DistributionChannel) AppDataRoot {
	t.Helper()
	root, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:                     channel,
		HostOS:                      HostOSWindows,
		WindowsPackageLocalDataRoot: `C:\Users\tester\AppData\Local\Packages\Riido_abc\LocalState`,
	})
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func mustDarwinStoreAppDataRoot(t *testing.T, channel DistributionChannel) AppDataRoot {
	t.Helper()
	root, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:            channel,
		HostOS:             HostOSDarwin,
		DarwinAppGroupRoot: "/Users/tester/Library/Group Containers/group.io.riido",
	})
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func hasReviewSurface(surfaces []string, wanted string) bool {
	return slices.Contains(surfaces, wanted)
}

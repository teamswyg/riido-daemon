package hostintegration

import (
	"slices"
	"testing"
)

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

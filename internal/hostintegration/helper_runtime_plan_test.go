package hostintegration

import "testing"

func TestResolveHelperRuntimePlanMSIXStoreBackgroundRequiresConsentAndReview(t *testing.T) {
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
			plan := resolveMSIXStorePlan(t, tt.consent, tt.review)
			if plan.BackgroundRule != HelperBackgroundConsentAndStoreReview {
				t.Fatalf("background rule = %q", plan.BackgroundRule)
			}
			if plan.BackgroundAllowed != tt.wantAllowed {
				t.Fatalf("background allowed = %v, want %v", plan.BackgroundAllowed, tt.wantAllowed)
			}
		})
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

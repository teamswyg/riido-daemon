package hostintegration

import (
	"testing"
)

func TestResolveHelperRuntimePlanMacAppStoreUsesSandboxedLoginItem(t *testing.T) {
	plan := resolveMacAppStorePlan(t, true, true)
	assertMacAppStorePlanShape(t, plan)
	assertStoreHelperSafetyPlan(t, plan)
	assertReviewSurfaces(t, plan, "helper-purpose-review-note", "service-management-login-item-consent")
}

func TestResolveHelperRuntimePlanMacAppStoreBackgroundRequiresConsentAndReview(t *testing.T) {
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
			plan := resolveMacAppStorePlan(t, tt.consent, tt.review)
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

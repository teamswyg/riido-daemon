package hostintegration

import "testing"

func TestResolveHelperRuntimePlanMSIXSideloadUsesPackagedBrokerWithoutStoreReview(t *testing.T) {
	plan := resolveMSIXSideloadPlan(t)

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

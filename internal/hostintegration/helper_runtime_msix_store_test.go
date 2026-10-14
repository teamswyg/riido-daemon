package hostintegration

import "testing"

func TestResolveHelperRuntimePlanMSIXStoreUsesFullTrustTrayNamedPipe(t *testing.T) {
	plan := resolveMSIXStorePlan(t, true, true)

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
	assertStoreHelperSafetyPlan(t, plan)
	assertReviewSurfaces(t, plan, "runfulltrust-review-note")
}

package hostintegration

import "testing"

func assertStoreHelperSafetyPlan(t *testing.T, plan HelperRuntimePlan) {
	t.Helper()
	if !plan.BackgroundAllowed {
		t.Fatal("store helper background should be allowed with consent and review approval")
	}
	if !plan.RequiresStoreReviewApproval {
		t.Fatal("store helper must require review approval")
	}
	if plan.ProviderCLIBundlingAllowed {
		t.Fatal("store helper plan must not allow bundled provider CLIs")
	}
	if plan.DirectLaunchAgentAllowed {
		t.Fatal("store helper plan must not allow direct LaunchAgent install")
	}
	if plan.WindowsServiceAllowed {
		t.Fatal("store helper plan must not allow Windows service install by default")
	}
	if plan.SharedLocationInstallAllowed {
		t.Fatal("store helper plan must not allow shared-location code install")
	}
	if plan.StandaloneCodeDownloadAllowed {
		t.Fatal("store helper plan must not allow standalone code download")
	}
	if plan.SelfUpdaterAllowed {
		t.Fatal("store helper plan must use Store-managed updates")
	}
}

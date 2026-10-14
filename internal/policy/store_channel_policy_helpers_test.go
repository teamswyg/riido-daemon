package policy_test

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func allStorePolicyChannels() []hostintegration.DistributionChannel {
	return []hostintegration.DistributionChannel{
		hostintegration.DistributionChannelDeveloperID,
		hostintegration.DistributionChannelMacAppStore,
		hostintegration.DistributionChannelMSIXSideload,
		hostintegration.DistributionChannelMSIXStore,
	}
}

func forbiddenStorePolicySurfaces() []policy.StoreSurface {
	return []policy.StoreSurface{
		policy.StoreSurfaceProviderCLIBundling,
		policy.StoreSurfaceSilentProviderAutoInstall,
		policy.StoreSurfaceExternalTCPListener,
		policy.StoreSurfaceArbitraryHomeScan,
	}
}

func evaluateStorePolicy(input policy.StoreChannelPolicyInput) policy.Decision {
	return policy.EvaluateStoreChannelPolicy(input)
}

func assertStorePolicyDeniedCode(
	t *testing.T,
	decision policy.Decision,
	code string,
) {
	t.Helper()
	if decision.Allowed || decision.Code != code {
		t.Fatalf("decision = %+v, want denied code %q", decision, code)
	}
}

func assertStorePolicyAllowed(t *testing.T, decision policy.Decision) {
	t.Helper()
	if !decision.Allowed {
		t.Fatalf("decision denied: %s", decision.Reason)
	}
}

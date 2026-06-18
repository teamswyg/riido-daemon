package hostintegration

import "testing"

func resolveMacAppStorePlan(t *testing.T, consent, review bool) HelperRuntimePlan {
	t.Helper()
	plan, err := ResolveHelperRuntimePlan(HelperRuntimePlanInput{
		Channel:             DistributionChannelMacAppStore,
		HostOS:              HostOSDarwin,
		AppDataRoot:         mustDarwinStoreAppDataRoot(t, DistributionChannelMacAppStore),
		Consent:             ConsentState{BackgroundHelper: consent},
		StoreReviewApproved: review,
		EndpointName:        "agentd",
	})
	if err != nil {
		t.Fatalf("ResolveHelperRuntimePlan: %v", err)
	}
	return plan
}

func resolveMSIXStorePlan(t *testing.T, consent, review bool) HelperRuntimePlan {
	t.Helper()
	plan, err := ResolveHelperRuntimePlan(HelperRuntimePlanInput{
		Channel:             DistributionChannelMSIXStore,
		HostOS:              HostOSWindows,
		AppDataRoot:         mustMSIXAppDataRoot(t, DistributionChannelMSIXStore),
		Consent:             ConsentState{BackgroundHelper: consent},
		StoreReviewApproved: review,
		EndpointName:        "agentd",
	})
	if err != nil {
		t.Fatalf("ResolveHelperRuntimePlan: %v", err)
	}
	return plan
}

func resolveMSIXSideloadPlan(t *testing.T) HelperRuntimePlan {
	t.Helper()
	plan, err := ResolveHelperRuntimePlan(HelperRuntimePlanInput{
		Channel:      DistributionChannelMSIXSideload,
		HostOS:       HostOSWindows,
		AppDataRoot:  mustMSIXAppDataRoot(t, DistributionChannelMSIXSideload),
		Consent:      ConsentState{BackgroundHelper: true},
		EndpointName: "agentd",
	})
	if err != nil {
		t.Fatalf("ResolveHelperRuntimePlan: %v", err)
	}
	return plan
}

package hostintegration

import "testing"

func TestEvaluateReviewDemoModeEnablesStoreManagedOfflineSurfaces(t *testing.T) {
	mode, err := EvaluateReviewDemoMode(ReviewDemoModeInput{
		Channel: DistributionChannelMSIXStore,
		Consent: ConsentState{
			ReviewDemoMode:  true,
			TelemetrySync:   true,
			ProviderExecute: nil,
		},
	})
	if err != nil {
		t.Fatalf("EvaluateReviewDemoMode: %v", err)
	}
	if !mode.Enabled {
		t.Fatal("review demo mode should be enabled")
	}
	if mode.ProviderExecutionAllowed {
		t.Fatal("review demo mode must not allow provider execution")
	}
	if mode.TelemetrySyncAllowed {
		t.Fatal("review demo mode must stay offline from telemetry sync")
	}
	if got, want := mode.Surfaces, []ReviewDemoSurface{
		ReviewDemoOnboarding,
		ReviewDemoProviderStatus,
		ReviewDemoWorkspaceGrant,
		ReviewDemoBackgroundConsent,
		ReviewDemoPrivacySettings,
		ReviewDemoLocalStatus,
	}; !sameReviewDemoSurfaces(got, want) {
		t.Fatalf("surfaces = %v, want %v", got, want)
	}
}

func TestEvaluateReviewDemoModeRequiresConsentForStoreManagedChannel(t *testing.T) {
	mode, err := EvaluateReviewDemoMode(ReviewDemoModeInput{Channel: DistributionChannelMacAppStore})
	if err == nil {
		t.Fatal("expected missing review-demo-mode consent error")
	}
	if mode.Enabled {
		t.Fatal("review demo mode should be disabled without consent")
	}
}

func TestEvaluateReviewDemoModeIgnoresNonStoreManagedChannel(t *testing.T) {
	mode, err := EvaluateReviewDemoMode(ReviewDemoModeInput{
		Channel: DistributionChannelDeveloperID,
		Consent: ConsentState{ReviewDemoMode: true},
	})
	if err != nil {
		t.Fatalf("EvaluateReviewDemoMode non-store: %v", err)
	}
	if mode.Enabled || len(mode.Surfaces) != 0 {
		t.Fatalf("non-store managed mode = %+v", mode)
	}
}

func sameReviewDemoSurfaces(got, want []ReviewDemoSurface) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

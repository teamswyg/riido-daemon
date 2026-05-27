package hostintegration

import (
	"errors"
	"fmt"
)

type ReviewDemoSurface string

const (
	ReviewDemoOnboarding        ReviewDemoSurface = "onboarding"
	ReviewDemoProviderStatus    ReviewDemoSurface = "provider-status"
	ReviewDemoWorkspaceGrant    ReviewDemoSurface = "workspace-grant"
	ReviewDemoBackgroundConsent ReviewDemoSurface = "background-consent"
	ReviewDemoPrivacySettings   ReviewDemoSurface = "privacy-settings"
	ReviewDemoLocalStatus       ReviewDemoSurface = "local-status"
)

type ReviewDemoModeInput struct {
	Channel DistributionChannel
	Consent ConsentState
}

type ReviewDemoMode struct {
	Channel                  DistributionChannel
	Enabled                  bool
	Surfaces                 []ReviewDemoSurface
	ProviderExecutionAllowed bool
	TelemetrySyncAllowed     bool
}

func EvaluateReviewDemoMode(input ReviewDemoModeInput) (ReviewDemoMode, error) {
	if !input.Channel.Valid() {
		return ReviewDemoMode{}, fmt.Errorf("unknown distribution channel %q", input.Channel)
	}
	mode := ReviewDemoMode{Channel: input.Channel}
	if !input.Channel.StoreManaged() {
		return mode, nil
	}
	if !input.Consent.ReviewDemoMode {
		return mode, errors.New("review demo mode requires review-demo-mode consent")
	}
	mode.Enabled = true
	mode.Surfaces = []ReviewDemoSurface{
		ReviewDemoOnboarding,
		ReviewDemoProviderStatus,
		ReviewDemoWorkspaceGrant,
		ReviewDemoBackgroundConsent,
		ReviewDemoPrivacySettings,
		ReviewDemoLocalStatus,
	}
	return mode, nil
}

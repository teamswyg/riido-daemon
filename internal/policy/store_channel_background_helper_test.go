package policy_test

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestStoreChannelPolicyBackgroundHelperRequiresConsentAndStoreReview(t *testing.T) {
	withoutConsent := evaluateStorePolicy(policy.StoreChannelPolicyInput{
		Channel: hostintegration.DistributionChannelDeveloperID,
		Surface: policy.StoreSurfaceBackgroundHelper,
	})
	assertStorePolicyDeniedCode(t, withoutConsent, "STORE_CHANNEL_REQUIRES_CONSENT")

	withoutReview := evaluateStorePolicy(policy.StoreChannelPolicyInput{
		Channel:                hostintegration.DistributionChannelMSIXStore,
		Surface:                policy.StoreSurfaceBackgroundHelper,
		ExplicitConsentGranted: true,
	})
	assertStorePolicyDeniedCode(t, withoutReview, "STORE_CHANNEL_REQUIRES_STORE_REVIEW_APPROVAL")

	withReview := evaluateStorePolicy(policy.StoreChannelPolicyInput{
		Channel:                hostintegration.DistributionChannelMSIXStore,
		Surface:                policy.StoreSurfaceBackgroundHelper,
		ExplicitConsentGranted: true,
		StoreReviewApproved:    true,
	})
	assertStorePolicyAllowed(t, withReview)
}

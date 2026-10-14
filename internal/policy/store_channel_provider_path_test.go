package policy_test

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestStoreChannelPolicyMacAppStoreProviderPathRequiresOSGrantAndStoreReview(t *testing.T) {
	withoutGrant := evaluateStorePolicy(policy.StoreChannelPolicyInput{
		Channel: hostintegration.DistributionChannelMacAppStore,
		Surface: policy.StoreSurfaceProviderCLIUserSelectedPath,
	})
	assertStorePolicyDeniedCode(t, withoutGrant, "STORE_CHANNEL_REQUIRES_OS_GRANT")

	withGrant := evaluateStorePolicy(policy.StoreChannelPolicyInput{
		Channel:        hostintegration.DistributionChannelMacAppStore,
		Surface:        policy.StoreSurfaceProviderCLIUserSelectedPath,
		OSGrantPresent: true,
	})
	assertStorePolicyDeniedCode(t, withGrant, "STORE_CHANNEL_REQUIRES_STORE_REVIEW_APPROVAL")

	withGrantAndReview := evaluateStorePolicy(policy.StoreChannelPolicyInput{
		Channel:             hostintegration.DistributionChannelMacAppStore,
		Surface:             policy.StoreSurfaceProviderCLIUserSelectedPath,
		OSGrantPresent:      true,
		StoreReviewApproved: true,
	})
	assertStorePolicyAllowed(t, withGrantAndReview)
}

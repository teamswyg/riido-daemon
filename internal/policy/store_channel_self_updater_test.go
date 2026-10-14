package policy_test

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestStoreChannelPolicyStoreManagedChannelsDenySelfUpdater(t *testing.T) {
	for _, channel := range []hostintegration.DistributionChannel{
		hostintegration.DistributionChannelMacAppStore,
		hostintegration.DistributionChannelMSIXStore,
	} {
		decision := evaluateStorePolicy(policy.StoreChannelPolicyInput{
			Channel: channel,
			Surface: policy.StoreSurfaceSelfUpdater,
		})
		if decision.Allowed {
			t.Fatalf("%s self-updater allowed, want store update mechanism only", channel)
		}
	}
}

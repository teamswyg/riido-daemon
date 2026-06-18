package policy_test

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestStoreChannelPolicyLocalIPCAllowedAcrossStoreChannels(t *testing.T) {
	for _, channel := range allStorePolicyChannels() {
		decision := evaluateStorePolicy(policy.StoreChannelPolicyInput{
			Channel: channel,
			Surface: policy.StoreSurfaceLocalIPC,
		})
		if !decision.Allowed {
			t.Fatalf("%s local IPC denied: %s", channel, decision.Reason)
		}
	}
}

package policy_test

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestStoreChannelPolicyAlwaysDeniesForbiddenSurfaces(t *testing.T) {
	for _, channel := range allStorePolicyChannels() {
		for _, surface := range forbiddenStorePolicySurfaces() {
			decision := evaluateStorePolicy(policy.StoreChannelPolicyInput{
				Channel: channel,
				Surface: surface,
			})
			if decision.Allowed {
				t.Fatalf("%s/%s allowed, want denied", channel, surface)
			}
			if decision.Code != "STORE_CHANNEL_SURFACE_FORBIDDEN" {
				t.Fatalf("%s/%s code = %q", channel, surface, decision.Code)
			}
		}
	}
}

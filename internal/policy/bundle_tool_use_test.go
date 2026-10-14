package policy

import "testing"

func TestParsePolicyBundleAllowsToolUseSurface(t *testing.T) {
	bundle := parsePolicyBundleFixture(t, `{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Host": {
				"allowed_surfaces": {
					"tool_use": ["tool:network-egress"]
				}
			}
		}
	}`)

	if !bundle.AllowsToolUse(TrustTierHost, ToolUseNetworkEgress) {
		t.Fatal("Host tier should allow explicitly-listed tool use surface")
	}
	allowed := EvaluateToolUseWithBundle(bundle, ToolUseInput{
		TrustTier: TrustTierHost,
		Surface:   ToolUseNetworkEgress,
	})
	if allowed.Action != ToolUseActionAllow {
		t.Fatalf("explicit tool use allow should pass: %+v", allowed)
	}
}

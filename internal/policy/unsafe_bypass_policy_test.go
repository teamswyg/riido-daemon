package policy

import "testing"

func TestEvaluateUnsafeBypassDeniesHostAndUnknown(t *testing.T) {
	for _, tier := range []TrustTier{"", TrustTierUnknown, TrustTierHost} {
		got := EvaluateUnsafeBypass(UnsafeBypassInput{
			TrustTier:    tier,
			Surface:      UnsafeBypassCursorYolo,
			BundleAllows: true,
		})
		if got.Allowed {
			t.Fatalf("tier %q must deny unsafe bypass: %+v", tier, got)
		}
	}
}

func TestEvaluateUnsafeBypassRequiresBundleAllowForIsolatedTier(t *testing.T) {
	denied := EvaluateUnsafeBypass(UnsafeBypassInput{
		TrustTier:    TrustTierIsolatedContainer,
		Surface:      UnsafeBypassClaudePermissions,
		BundleAllows: false,
	})
	if denied.Allowed || denied.Code != "UNSAFE_BYPASS_NOT_IN_POLICY_BUNDLE" {
		t.Fatalf("isolated tier without bundle allow should deny: %+v", denied)
	}

	allowed := EvaluateUnsafeBypass(UnsafeBypassInput{
		TrustTier:    TrustTierEphemeralVM,
		Surface:      UnsafeBypassClaudePermissions,
		BundleAllows: true,
	})
	if !allowed.Allowed {
		t.Fatalf("isolated tier with bundle allow should allow: %+v", allowed)
	}
}

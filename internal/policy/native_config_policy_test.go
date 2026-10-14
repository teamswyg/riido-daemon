package policy

import "testing"

func TestEvaluateNativeConfigHookRequiresKnownTierAndBundleAllow(t *testing.T) {
	for _, tier := range []TrustTier{"", TrustTierUnknown, TrustTier("MoonBase")} {
		got := EvaluateNativeConfigHook(NativeConfigHookInput{
			TrustTier:    tier,
			Surface:      NativeConfigHookClaudeCommandAudit,
			BundleAllows: true,
		})
		if got.Allowed {
			t.Fatalf("tier %q must deny native config hook: %+v", tier, got)
		}
	}

	denied := EvaluateNativeConfigHook(NativeConfigHookInput{
		TrustTier:    TrustTierHost,
		Surface:      NativeConfigHookClaudeCommandAudit,
		BundleAllows: false,
	})
	if denied.Allowed || denied.Code != "NATIVE_CONFIG_HOOK_NOT_IN_POLICY_BUNDLE" {
		t.Fatalf("Host tier without bundle allow should deny: %+v", denied)
	}

	allowed := EvaluateNativeConfigHook(NativeConfigHookInput{
		TrustTier:    TrustTierHost,
		Surface:      NativeConfigHookClaudeCommandAudit,
		BundleAllows: true,
	})
	if !allowed.Allowed {
		t.Fatalf("Host tier with bundle allow should allow: %+v", allowed)
	}
}

func TestEvaluateNativeConfigFileRequiresKnownTierAndBundleAllow(t *testing.T) {
	for _, tier := range []TrustTier{"", TrustTierUnknown, TrustTier("MoonBase")} {
		got := EvaluateNativeConfigFile(NativeConfigFileInput{
			TrustTier:    tier,
			Surface:      NativeConfigFileCodexTaskScopedHome,
			BundleAllows: true,
		})
		if got.Allowed {
			t.Fatalf("tier %q must deny native config file: %+v", tier, got)
		}
	}

	denied := EvaluateNativeConfigFile(NativeConfigFileInput{
		TrustTier:    TrustTierHost,
		Surface:      NativeConfigFileCodexTaskScopedHome,
		BundleAllows: false,
	})
	if denied.Allowed || denied.Code != "NATIVE_CONFIG_FILE_NOT_IN_POLICY_BUNDLE" {
		t.Fatalf("Host tier without bundle allow should deny: %+v", denied)
	}

	allowed := EvaluateNativeConfigFile(NativeConfigFileInput{
		TrustTier:    TrustTierHost,
		Surface:      NativeConfigFileCodexTaskScopedHome,
		BundleAllows: true,
	})
	if !allowed.Allowed {
		t.Fatalf("Host tier with bundle allow should allow: %+v", allowed)
	}
}

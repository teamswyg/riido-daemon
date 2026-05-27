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

func TestEvaluateToolUseBranchesByPolicyAndApproval(t *testing.T) {
	allowed := EvaluateToolUse(ToolUseInput{
		TrustTier:    TrustTierHost,
		Surface:      ToolUseNetworkEgress,
		BundleAllows: true,
	})
	if allowed.Action != ToolUseActionAllow {
		t.Fatalf("explicit tool use allow = %+v, want allow", allowed)
	}

	approval := EvaluateToolUse(ToolUseInput{
		TrustTier:              TrustTierHost,
		Surface:                ToolUseProtectedPathWrite,
		HumanApprovalAvailable: true,
	})
	if approval.Action != ToolUseActionRequireApproval || approval.Code != "TOOL_USE_REQUIRES_APPROVAL" {
		t.Fatalf("missing tool use allow with approval = %+v, want require approval", approval)
	}

	blocked := EvaluateToolUse(ToolUseInput{
		TrustTier: TrustTierHost,
		Surface:   ToolUseSecretExposure,
	})
	if blocked.Action != ToolUseActionInterruptAndBlock || blocked.Code != "TOOL_USE_NOT_IN_POLICY_BUNDLE" {
		t.Fatalf("missing tool use allow without approval = %+v, want interrupt-and-block", blocked)
	}
}

func TestEvaluateToolUseBlocksUnknownTierAndSurface(t *testing.T) {
	for _, tier := range []TrustTier{"", TrustTierUnknown, TrustTier("MoonBase")} {
		got := EvaluateToolUse(ToolUseInput{
			TrustTier:    tier,
			Surface:      ToolUseNetworkEgress,
			BundleAllows: true,
		})
		if got.Action != ToolUseActionInterruptAndBlock || got.Code != "TOOL_USE_UNKNOWN_TRUST_TIER" {
			t.Fatalf("tier %q tool use decision = %+v, want unknown tier block", tier, got)
		}
	}

	got := EvaluateToolUse(ToolUseInput{
		TrustTier:    TrustTierHost,
		Surface:      ToolUseSurface("tool:ghost"),
		BundleAllows: true,
	})
	if got.Action != ToolUseActionInterruptAndBlock || got.Code != "TOOL_USE_UNKNOWN_SURFACE" {
		t.Fatalf("unknown surface tool use decision = %+v, want unknown surface block", got)
	}
}

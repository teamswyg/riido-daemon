package policy

import (
	"strings"
	"testing"
)

func TestParsePolicyBundleAllowsIsolatedUnsafeBypass(t *testing.T) {
	bundle := parsePolicyBundleFixture(t, `{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"IsolatedContainer": {
				"allowed_surfaces": {
					"unsafe_bypass": [
						"claude:bypassPermissions",
						"codex:--yolo"
					]
				}
			},
			"Host": {
				"allowed_surfaces": {}
			}
		}
	}`)

	if bundle.Version != "policy-bundle.test.v1" {
		t.Fatalf("version = %q", bundle.Version)
	}
	if !bundle.AllowsUnsafeBypass(TrustTierIsolatedContainer, UnsafeBypassCodexYolo) {
		t.Fatal("isolated tier should allow explicitly-listed Codex yolo surface")
	}
	if bundle.AllowsUnsafeBypass(TrustTierHost, UnsafeBypassCodexYolo) {
		t.Fatal("Host tier must not allow unsafe bypass")
	}

	allowed := EvaluateUnsafeBypassWithBundle(bundle, UnsafeBypassInput{
		TrustTier: TrustTierIsolatedContainer,
		Surface:   UnsafeBypassCodexYolo,
	})
	if !allowed.Allowed {
		t.Fatalf("explicit isolated bundle allow should pass: %+v", allowed)
	}
}

func TestParsePolicyBundleAllowsHostNativeConfigHook(t *testing.T) {
	bundle := parsePolicyBundleFixture(t, `{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Host": {
				"allowed_surfaces": {
					"native_config_hooks": [
						"claude:command-hooks:audit"
					]
				}
			}
		}
	}`)

	if !bundle.AllowsNativeConfigHook(TrustTierHost, NativeConfigHookClaudeCommandAudit) {
		t.Fatal("Host tier should allow explicitly-listed audit hook surface")
	}
	allowed := EvaluateNativeConfigHookWithBundle(bundle, NativeConfigHookInput{
		TrustTier: TrustTierHost,
		Surface:   NativeConfigHookClaudeCommandAudit,
	})
	if !allowed.Allowed {
		t.Fatalf("explicit native config hook allow should pass: %+v", allowed)
	}
}

func TestParsePolicyBundleAllowsHostNativeConfigFile(t *testing.T) {
	bundle := parsePolicyBundleFixture(t, `{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Host": {
				"allowed_surfaces": {
					"native_config_files": [
						"codex:config-home:task-scoped"
					]
				}
			}
		}
	}`)

	if !bundle.AllowsNativeConfigFile(TrustTierHost, NativeConfigFileCodexTaskScopedHome) {
		t.Fatal("Host tier should allow explicitly-listed Codex task-scoped config home surface")
	}
	allowed := EvaluateNativeConfigFileWithBundle(bundle, NativeConfigFileInput{
		TrustTier: TrustTierHost,
		Surface:   NativeConfigFileCodexTaskScopedHome,
	})
	if !allowed.Allowed {
		t.Fatalf("explicit native config file allow should pass: %+v", allowed)
	}
}

func TestParsePolicyBundleAllowsToolUseSurface(t *testing.T) {
	bundle := parsePolicyBundleFixture(t, `{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Host": {
				"allowed_surfaces": {
					"tool_use": [
						"tool:network-egress"
					]
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

func TestDefaultLocalPolicyBundleAllowsOnlyClaudeNativeConfigSurface(t *testing.T) {
	bundle := DefaultLocalPolicyBundle()
	if err := bundle.Validate(); err != nil {
		t.Fatal(err)
	}
	if bundle.Version != DefaultLocalPolicyBundleVersion {
		t.Fatalf("default bundle version = %q", bundle.Version)
	}
	if !bundle.AllowsNativeConfigHook(TrustTierHost, NativeConfigHookClaudeCommandAudit) {
		t.Fatal("local default should allow Claude audit hook materialization")
	}
	if bundle.AllowsNativeConfigFile(TrustTierHost, NativeConfigFileCodexTaskScopedHome) {
		t.Fatal("local default must not materialize Codex task-scoped config home")
	}
	if bundle.AllowsUnsafeBypass(TrustTierHost, UnsafeBypassClaudePermissions) {
		t.Fatal("local default must not allow unsafe bypass on Host")
	}
	if bundle.AllowsToolUse(TrustTierHost, ToolUseNetworkEgress) {
		t.Fatal("local default must not silently allow tool-use risk surfaces")
	}
}

func TestParsePolicyBundleRejectsHostUnsafeBypass(t *testing.T) {
	_, err := ParsePolicyBundleJSON([]byte(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Host": {
				"allowed_surfaces": {
					"unsafe_bypass": ["cursor:--yolo"]
				}
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "Host") {
		t.Fatalf("expected Host unsafe bypass rejection, got %v", err)
	}
}

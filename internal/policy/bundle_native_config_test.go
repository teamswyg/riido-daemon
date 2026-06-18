package policy

import "testing"

func TestParsePolicyBundleAllowsHostNativeConfigHook(t *testing.T) {
	bundle := parsePolicyBundleFixture(t, `{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Host": {
				"allowed_surfaces": {
					"native_config_hooks": ["claude:command-hooks:audit"]
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
					"native_config_files": ["codex:config-home:task-scoped"]
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

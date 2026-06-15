package policy

import (
	"strings"
	"testing"
)

func TestParsePolicyBundleRejectsUnknownSurface(t *testing.T) {
	_, err := ParsePolicyBundleJSON([]byte(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"EphemeralVM": {
				"allowed_surfaces": {
					"unsafe_bypass": ["cursor:--ghost"]
				}
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "unknown unsafe bypass surface") {
		t.Fatalf("expected unknown surface rejection, got %v", err)
	}
}

func TestParsePolicyBundleRejectsUnknownNativeConfigHook(t *testing.T) {
	_, err := ParsePolicyBundleJSON([]byte(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Host": {
				"allowed_surfaces": {
					"native_config_hooks": ["claude:command-hooks:blocking"]
				}
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "unknown native config hook surface") {
		t.Fatalf("expected unknown native config hook rejection, got %v", err)
	}
}

func TestParsePolicyBundleRejectsUnknownNativeConfigFile(t *testing.T) {
	_, err := ParsePolicyBundleJSON([]byte(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Host": {
				"allowed_surfaces": {
					"native_config_files": ["codex:config-home:global"]
				}
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "unknown native config file surface") {
		t.Fatalf("expected unknown native config file rejection, got %v", err)
	}
}

func TestParsePolicyBundleRejectsUnknownToolUseSurface(t *testing.T) {
	_, err := ParsePolicyBundleJSON([]byte(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Host": {
				"allowed_surfaces": {
					"tool_use": ["tool:teleport"]
				}
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "unknown tool use surface") {
		t.Fatalf("expected unknown tool use surface rejection, got %v", err)
	}
}

func TestParsePolicyBundleRejectsUnknownTierNativeConfigHook(t *testing.T) {
	_, err := ParsePolicyBundleJSON([]byte(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Unknown": {
				"allowed_surfaces": {
					"native_config_hooks": ["claude:command-hooks:audit"]
				}
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "Unknown") {
		t.Fatalf("expected Unknown native config hook rejection, got %v", err)
	}
}

func TestParsePolicyBundleRejectsUnknownTierNativeConfigFile(t *testing.T) {
	_, err := ParsePolicyBundleJSON([]byte(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Unknown": {
				"allowed_surfaces": {
					"native_config_files": ["codex:config-home:task-scoped"]
				}
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "Unknown") {
		t.Fatalf("expected Unknown native config file rejection, got %v", err)
	}
}

func TestParsePolicyBundleRejectsUnknownTierToolUse(t *testing.T) {
	_, err := ParsePolicyBundleJSON([]byte(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Unknown": {
				"allowed_surfaces": {
					"tool_use": ["tool:network-egress"]
				}
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "Unknown") {
		t.Fatalf("expected Unknown tool use rejection, got %v", err)
	}
}

func TestParsePolicyBundleRejectsDuplicateNativeConfigFile(t *testing.T) {
	_, err := ParsePolicyBundleJSON([]byte(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Host": {
				"allowed_surfaces": {
					"native_config_files": [
						"codex:config-home:task-scoped",
						"codex:config-home:task-scoped"
					]
				}
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "duplicate native config file surface") {
		t.Fatalf("expected duplicate native config file rejection, got %v", err)
	}
}

func TestParsePolicyBundleRejectsDuplicateToolUseSurface(t *testing.T) {
	_, err := ParsePolicyBundleJSON([]byte(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			"Host": {
				"allowed_surfaces": {
					"tool_use": [
						"tool:network-egress",
						"tool:network-egress"
					]
				}
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "duplicate tool use surface") {
		t.Fatalf("expected duplicate tool use rejection, got %v", err)
	}
}

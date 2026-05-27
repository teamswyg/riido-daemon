package policy

import (
	"os"
	"path/filepath"
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

func TestDefaultLocalPolicyBundleAllowsOnlyNativeConfigSurfaces(t *testing.T) {
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
	if !bundle.AllowsNativeConfigFile(TrustTierHost, NativeConfigFileCodexTaskScopedHome) {
		t.Fatal("local default should allow Codex task-scoped config home materialization")
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

func TestParsePolicyBundleRejectsUnknownField(t *testing.T) {
	_, err := ParsePolicyBundleJSON([]byte(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {},
		"surprise": true
	}`))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field rejection, got %v", err)
	}
}

func TestParsePolicyBundleRejectsMissingTrustTierPolicies(t *testing.T) {
	_, err := ParsePolicyBundleJSON([]byte(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z"
	}`))
	if err == nil || !strings.Contains(err.Error(), "trust_tier_policies") {
		t.Fatalf("expected missing trust_tier_policies rejection, got %v", err)
	}
}

func TestLoadPolicyBundleFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy-bundle.riido.json")
	if err := os.WriteFile(path, []byte(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.file.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {}
	}`), 0o600); err != nil {
		t.Fatal(err)
	}

	bundle, err := LoadPolicyBundleFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Version != "policy-bundle.file.v1" {
		t.Fatalf("version = %q", bundle.Version)
	}
}

func parsePolicyBundleFixture(t *testing.T, src string) PolicyBundle {
	t.Helper()
	bundle, err := ParsePolicyBundleJSON([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

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
					"unsafe_bypass": ["claude:bypassPermissions", "codex:--yolo"]
				}
			},
			"Host": {"allowed_surfaces": {}}
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

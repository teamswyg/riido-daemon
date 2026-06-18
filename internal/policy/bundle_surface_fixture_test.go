package policy

import (
	"fmt"
	"testing"
)

func parseBundleSurfaceError(t *testing.T, tier, surfaces string) string {
	t.Helper()

	body := fmt.Sprintf(`{
		"schema_version": "riido-policy-bundle.v1",
		"version": "policy-bundle.test.v1",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {
			%q: {
				"allowed_surfaces": {
					%s
				}
			}
		}
	}`, tier, surfaces)
	_, err := ParsePolicyBundleJSON([]byte(body))
	if err == nil {
		t.Fatal("expected policy bundle parse error")
	}
	return err.Error()
}

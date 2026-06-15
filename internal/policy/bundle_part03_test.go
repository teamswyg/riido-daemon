package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

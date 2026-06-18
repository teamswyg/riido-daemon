package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDaemonSettingsLoadsPolicyBundleFile(t *testing.T) {
	path := writePolicyBundleFile(t, "policy-bundle.file.v1")
	env := map[string]string{envPolicyBundlePath: path}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.PolicyBundle != "policy-bundle.file.v1" || settings.PolicyBundlePath != path {
		t.Fatalf("policy bundle settings = %+v", settings)
	}
	if settings.PolicyBundleDoc.Version != "policy-bundle.file.v1" {
		t.Fatalf("policy bundle doc = %+v", settings.PolicyBundleDoc)
	}
}

func TestLoadDaemonSettingsRejectsPolicyBundleVersionMismatch(t *testing.T) {
	path := writePolicyBundleFile(t, "policy-bundle.file.v1")
	env := map[string]string{
		envPolicyBundlePath: path,
		envPolicyBundle:     "policy-bundle.env.v1",
	}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected policy bundle version mismatch error")
	}
}

func TestLoadDaemonSettingsRejectsInvalidPolicyBundleFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy-bundle.riido.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"wrong"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{envPolicyBundlePath: path}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected invalid policy bundle file error")
	}
}

func writePolicyBundleFile(t *testing.T, version string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "policy-bundle.riido.json")
	body := `{
		"schema_version": "riido-policy-bundle.v1",
		"version": "` + version + `",
		"effective_since": "2026-05-27T00:00:00Z",
		"trust_tier_policies": {}
	}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

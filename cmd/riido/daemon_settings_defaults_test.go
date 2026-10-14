package main

import (
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestLoadDaemonSettingsDefaultWorkdirRoot(t *testing.T) {
	old := binaryVersion
	binaryVersion = "v-default"
	t.Cleanup(func() { binaryVersion = old })

	settings, err := loadDaemonSettingsFromEnvWithHome(
		func(string) string { return "" },
		func() (string, error) { return "host", nil },
		func() (string, error) { return "/Users/tester", nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/Users/tester", "Library", "Application Support", "riido", "workspaces")
	if settings.WorkdirRoot != want {
		t.Fatalf("workdir root = %q, want %q", settings.WorkdirRoot, want)
	}
	if settings.PolicyBundle != "policy-bundle.local.v0" {
		t.Fatalf("policy bundle = %q", settings.PolicyBundle)
	}
	if settings.PolicyBundleDoc.Version != policy.DefaultLocalPolicyBundleVersion ||
		!settings.PolicyBundleDoc.AllowsNativeConfigHook(policy.TrustTierHost, policy.NativeConfigHookClaudeCommandAudit) ||
		settings.PolicyBundleDoc.AllowsNativeConfigFile(policy.TrustTierHost, policy.NativeConfigFileCodexTaskScopedHome) {
		t.Fatalf("default policy bundle doc = %+v", settings.PolicyBundleDoc)
	}
	if settings.DaemonVersion != "riido-agentd v-default" ||
		settings.WorkdirRetention != 0 ||
		settings.WorkdirCleanupEvery != 0 {
		t.Fatalf("default settings mismatch: %+v", settings)
	}
}

package main

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestLoadDaemonSettingsReadsWorkdirAndPolicy(t *testing.T) {
	settings := loadDaemonSettingsForTest(t, fullDaemonSettingsEnv())
	if settings.WorkdirRoot != "/tmp/riido-workspaces" {
		t.Fatalf("workdir root mismatch: %+v", settings)
	}
	if settings.PolicyBundle != "policy-bundle.test.v1" {
		t.Fatalf("policy bundle mismatch: %+v", settings)
	}
	doc := settings.PolicyBundleDoc
	if doc.Version != "policy-bundle.test.v1" ||
		!doc.AllowsNativeConfigHook(policy.TrustTierHost, policy.NativeConfigHookClaudeCommandAudit) ||
		doc.AllowsNativeConfigFile(policy.TrustTierHost, policy.NativeConfigFileCodexTaskScopedHome) {
		t.Fatalf("default policy bundle doc mismatch: %+v", doc)
	}
}

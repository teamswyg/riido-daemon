package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestLoadDaemonSettingsAcceptsSaaSControlPlane(t *testing.T) {
	env := map[string]string{
		envSaaSURL:      "https://api.riido.ai",
		envDeviceID:     "device-1",
		envDeviceSecret: "rdev-secret",
	}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.SaaSURL != "https://api.riido.ai" || settings.DeviceID != "device-1" || settings.DeviceSecret != "rdev-secret" {
		t.Fatalf("saas settings = %+v", settings)
	}
	if settings.DaemonID != "device-1" {
		t.Fatalf("SaaS device credential daemon id = %q, want device principal id", settings.DaemonID)
	}
}

func TestLoadDaemonSettingsKeepsExplicitDaemonIDForSaaSControlPlane(t *testing.T) {
	env := map[string]string{
		envDaemonID:     "explicit-daemon",
		envSaaSURL:      "https://api.riido.ai",
		envDeviceID:     "device-1",
		envDeviceSecret: "rdev-secret",
	}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.DaemonID != "explicit-daemon" {
		t.Fatalf("explicit daemon id = %q", settings.DaemonID)
	}
}

func TestLoadDaemonSettingsIgnoresLegacySaaSEnvsWithDeviceCredential(t *testing.T) {
	env := map[string]string{
		envSaaSURL:          "https://api.riido.ai",
		envDeviceID:         "device-1",
		envDeviceSecret:     "rdev-secret",
		"RIIDO_SAAS_AGENTS": "jykim1:codex",
		"RIIDO_SAAS_TOKEN":  "secret",
	}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.DeviceID != "device-1" || settings.DeviceSecret != "rdev-secret" {
		t.Fatalf("device credential settings = %+v", settings)
	}
}

func TestLoadDaemonSettingsAcceptsDynamicSaaSDeviceCredential(t *testing.T) {
	env := map[string]string{
		envSaaSURL:      "https://api.riido.ai",
		envDeviceID:     "device-1",
		envDeviceSecret: "rdev-secret",
	}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.SaaSURL != "https://api.riido.ai" || settings.DeviceID != "device-1" || settings.DeviceSecret != "rdev-secret" {
		t.Fatalf("dynamic device credential settings = %+v", settings)
	}
}

func TestLoadDaemonSettingsRejectsIncompleteSaaSDeviceCredential(t *testing.T) {
	env := map[string]string{
		envSaaSURL:  "https://api.riido.ai",
		envDeviceID: "device-1",
	}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil || !strings.Contains(err.Error(), envDeviceID) || !strings.Contains(err.Error(), envDeviceSecret) {
		t.Fatalf("expected incomplete device credential error, got %v", err)
	}
}

func TestLoadDaemonSettingsRejectsInvalidPollInterval(t *testing.T) {
	env := map[string]string{envDaemonPollIntervalSeconds: "0"}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected invalid poll interval error")
	}
}

func TestLoadDaemonSettingsRejectsIdlePollBelowActivePoll(t *testing.T) {
	env := map[string]string{
		envDaemonPollIntervalSeconds:     "10",
		envDaemonIdlePollIntervalSeconds: "3",
	}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected idle poll interval below active poll interval error")
	}
}

func TestLoadDaemonSettingsRejectsSaaSWithoutDeviceCredential(t *testing.T) {
	env := map[string]string{envSaaSURL: "https://api.riido.ai"}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected SaaS URL without device credential error")
	}
}

func TestLoadDaemonSettingsRejectsSaaSWithTaskDBSource(t *testing.T) {
	env := map[string]string{
		envSaaSURL:          "https://api.riido.ai",
		envDeviceID:         "device-1",
		envDeviceSecret:     "rdev-secret",
		envTaskDBSourcePath: "/tmp/task-db.json",
	}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected SaaS and task DB conflict")
	}
}

func TestLoadDaemonSettingsDefaultWorkdirRoot(t *testing.T) {
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
	if settings.DaemonVersion != "riido-agentd v0.0.0" {
		t.Fatalf("daemon version = %q", settings.DaemonVersion)
	}
	if settings.WorkdirRetention != 0 || settings.WorkdirCleanupEvery != 0 {
		t.Fatalf("workdir cleanup should default disabled: %+v", settings)
	}
}

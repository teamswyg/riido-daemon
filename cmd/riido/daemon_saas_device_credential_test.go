package main

import "testing"

func TestLoadDaemonSettingsAcceptsSaaSControlPlane(t *testing.T) {
	env := saasDeviceCredentialEnv()
	settings := loadDaemonSettingsForTest(t, env)
	if settings.SaaSURL != "https://api.riido.ai" ||
		settings.DeviceID != "device-1" ||
		settings.DeviceSecret != "rdev-secret" {
		t.Fatalf("saas settings = %+v", settings)
	}
	if settings.DaemonID != "device-1" {
		t.Fatalf("SaaS device credential daemon id = %q, want device principal id", settings.DaemonID)
	}
}

func TestLoadDaemonSettingsKeepsExplicitDaemonIDForSaaSControlPlane(t *testing.T) {
	env := saasDeviceCredentialEnv()
	env[envDaemonID] = "explicit-daemon"
	settings := loadDaemonSettingsForTest(t, env)
	if settings.DaemonID != "explicit-daemon" {
		t.Fatalf("explicit daemon id = %q", settings.DaemonID)
	}
}

func TestLoadDaemonSettingsIgnoresLegacySaaSEnvsWithDeviceCredential(t *testing.T) {
	env := saasDeviceCredentialEnv()
	env["RIIDO_SAAS_AGENTS"] = "jykim1:codex"
	env["RIIDO_SAAS_TOKEN"] = "secret"
	settings := loadDaemonSettingsForTest(t, env)
	if settings.DeviceID != "device-1" || settings.DeviceSecret != "rdev-secret" {
		t.Fatalf("device credential settings = %+v", settings)
	}
}

func TestLoadDaemonSettingsAcceptsDynamicSaaSDeviceCredential(t *testing.T) {
	settings := loadDaemonSettingsForTest(t, saasDeviceCredentialEnv())
	if settings.SaaSURL != "https://api.riido.ai" ||
		settings.DeviceID != "device-1" ||
		settings.DeviceSecret != "rdev-secret" {
		t.Fatalf("dynamic device credential settings = %+v", settings)
	}
}

func saasDeviceCredentialEnv() map[string]string {
	return map[string]string{
		envSaaSURL:      "https://api.riido.ai",
		envDeviceID:     "device-1",
		envDeviceSecret: "rdev-secret",
	}
}

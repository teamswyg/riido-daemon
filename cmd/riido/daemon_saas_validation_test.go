package main

import (
	"strings"
	"testing"
)

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
	env := saasDeviceCredentialEnv()
	env[envTaskDBSourcePath] = "/tmp/task-db.json"
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected SaaS and task DB conflict")
	}
}

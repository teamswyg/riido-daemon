package main

import (
	"testing"
	"time"
)

func TestLoadDaemonSettingsRejectsInvalidWorkspaceCount(t *testing.T) {
	env := map[string]string{envWorkspaceCount: "nope"}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected invalid workspace count error")
	}
}

func TestLoadDaemonSettingsDefaultsWorkdirCleanupInterval(t *testing.T) {
	env := map[string]string{envWorkdirRetentionSeconds: "7200"}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.WorkdirRetention != 2*time.Hour || settings.WorkdirCleanupEvery != time.Hour {
		t.Fatalf("workdir cleanup settings = %+v", settings)
	}
}

func TestLoadDaemonSettingsRejectsCleanupIntervalWithoutRetention(t *testing.T) {
	env := map[string]string{envWorkdirCleanupIntervalSeconds: "60"}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil {
		t.Fatal("expected cleanup interval without retention error")
	}
}

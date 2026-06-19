package main

import (
	"strings"
	"testing"
)

func TestLoadDaemonSettingsEnablesPprofForDevelopmentProfile(t *testing.T) {
	env := map[string]string{envDaemonProfile: "development"}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.PprofAddr != defaultDevelopmentPprofAddr {
		t.Fatalf("pprof addr = %q, want %q", settings.PprofAddr, defaultDevelopmentPprofAddr)
	}
}

func TestLoadDaemonSettingsKeepsPprofDisabledByDefault(t *testing.T) {
	settings, err := loadDaemonSettingsFromEnv(func(string) string { return "" }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.PprofAddr != "" {
		t.Fatalf("pprof addr = %q, want disabled", settings.PprofAddr)
	}
}

func TestLoadDaemonSettingsAllowsPprofDevelopmentOverrideOff(t *testing.T) {
	env := map[string]string{
		envDaemonProfile:   "development",
		envDaemonPprofAddr: "off",
	}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.PprofAddr != "" {
		t.Fatalf("pprof addr = %q, want disabled", settings.PprofAddr)
	}
}

func TestLoadDaemonSettingsRejectsNonLoopbackPprofAddr(t *testing.T) {
	externalAddr := strings.Join([]string{"0", "0", "0", "0"}, ".") + ":6061"
	env := map[string]string{envDaemonPprofAddr: externalAddr}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil || !strings.Contains(err.Error(), envDaemonPprofAddr) {
		t.Fatalf("expected %s validation error, got %v", envDaemonPprofAddr, err)
	}
}

package main

import "testing"

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

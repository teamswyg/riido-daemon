package main

import "testing"

// RIIDO_RUNTIME_MAX_CONCURRENT controls how many runtime sessions a provider can
// run at once. Default is 4 (so a single agent isn't limited to one task at a
// time); an explicit value overrides it.
func TestRuntimeMaxConcurrentDefaultAndOverride(t *testing.T) {
	hostname := func() (string, error) { return "host", nil }

	def, err := loadDaemonSettingsFromEnv(func(string) string { return "" }, hostname)
	if err != nil {
		t.Fatalf("default load: %v", err)
	}
	if def.RuntimeMaxConcurrent != 4 {
		t.Fatalf("default RuntimeMaxConcurrent = %d, want 4", def.RuntimeMaxConcurrent)
	}

	env := map[string]string{envRuntimeMaxConcurrent: "8"}
	got, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, hostname)
	if err != nil {
		t.Fatalf("override load: %v", err)
	}
	if got.RuntimeMaxConcurrent != 8 {
		t.Fatalf("override RuntimeMaxConcurrent = %d, want 8", got.RuntimeMaxConcurrent)
	}
}

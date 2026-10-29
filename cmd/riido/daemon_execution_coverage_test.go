package main

import (
	"strings"
	"testing"
	"time"
)

func TestDaemonRejectsUnnegotiatedCoveredCredential(t *testing.T) {
	env := saasDeviceCredentialEnv()
	env[envDeviceID] = "dcov_fixture"
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) { return "host", nil })
	if err == nil || !strings.Contains(err.Error(), "RIIDO_EXECUTION_COVERAGE_GENERATION") {
		t.Fatal("covered device started without its server generation")
	}
}

func TestDaemonPreservesCoverageEnrollmentInPlane(t *testing.T) {
	env := saasDeviceCredentialEnv()
	env[envDeviceID] = "dcov_fixture"
	env[envExecutionCoverageGeneration] = "cov_fixture.Exact-Case_1"
	settings := loadDaemonSettingsForTest(t, env)
	if settings.ExecutionCoverageGeneration != env[envExecutionCoverageGeneration] {
		t.Fatal("server generation changed during settings load")
	}
	source, _, _, err := buildDaemonControlPlane(settings, time.Time{})
	if err != nil {
		t.Fatal("covered daemon settings did not reach the transport")
	}
	if closer, ok := source.(interface{ Close() }); ok {
		closer.Close()
	}
}

func TestDaemonCoverageDoesNotNormalizeMalformedEnvironment(t *testing.T) {
	for _, value := range []string{" cov_fixture", "cov_fixture\n", "cov fixture"} {
		env := saasDeviceCredentialEnv()
		env[envDeviceID] = "dcov_fixture"
		env[envExecutionCoverageGeneration] = value
		if _, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) { return "host", nil }); err == nil {
			t.Fatal("malformed enrollment metadata was normalized or accepted")
		}
	}
}

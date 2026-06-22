package main

import "testing"

func TestIntegrationSkippedDetectsGoTestSkip(t *testing.T) {
	out := "=== RUN   TestIntegration\n--- SKIP: TestIntegration (0.01s)\nPASS\n"
	if !integrationSkipped(out) {
		t.Fatal("expected skip output to be classified as skipped")
	}
}

func TestIntegrationSkippedIgnoresPassingRun(t *testing.T) {
	out := "=== RUN   TestIntegration\n--- PASS: TestIntegration (0.01s)\nPASS\n"
	if integrationSkipped(out) {
		t.Fatal("passing integration must not be classified as skipped")
	}
}

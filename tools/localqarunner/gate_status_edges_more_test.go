package main

import (
	"testing"
	"time"
)

func TestCoverageRowExpiredIgnoresEmptyMalformedAndFutureDates(t *testing.T) {
	now := time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC)
	for name, row := range map[string]runCoverageRow{
		"empty":     {},
		"malformed": {ExpiresAt: "tomorrow"},
		"future":    {ExpiresAt: now.Add(time.Second).Format(time.RFC3339)},
	} {
		if coverageRowExpired(row, now) {
			t.Fatalf("%s row should not be expired", name)
		}
	}
	for _, expiresAt := range []string{
		now.Format(time.RFC3339),
		now.Add(-time.Second).Format(time.RFC3339),
	} {
		if !coverageRowExpired(runCoverageRow{ExpiresAt: expiresAt}, now) {
			t.Fatalf("row expiring at %s should be expired", expiresAt)
		}
	}
}

func TestDeploymentBlockerHelpersSkipEmptyPassedAndZeroCounts(t *testing.T) {
	if got := appendRunStatusBlocker(nil, ""); len(got) != 0 {
		t.Fatalf("empty run status should not block: %#v", got)
	}
	if got := appendRunStatusBlocker(nil, statusPassed); len(got) != 0 {
		t.Fatalf("passed run status should not block: %#v", got)
	}
	if got := appendExpiredCoverageBlocker(nil, 0); len(got) != 0 {
		t.Fatalf("zero expired coverage should not block: %#v", got)
	}
	got := appendRunStatusBlocker(nil, statusFailed)
	if len(got) != 1 || got[0].Code != "run_status_not_passed" {
		t.Fatalf("failed run should block: %#v", got)
	}
}

func TestMergeCoverageStatusPrioritizesFailedThenPartial(t *testing.T) {
	for _, pair := range [][2]string{
		{statusFailed, statusPassed},
		{statusPassed, statusFailed},
	} {
		if got := mergeCoverageStatus(pair[0], pair[1]); got != statusFailed {
			t.Fatalf("failed should win for %#v, got %s", pair, got)
		}
	}
	if got := mergeCoverageStatus("", statusPassed); got != statusPassed {
		t.Fatalf("empty current should default to passed, got %s", got)
	}
	if got := mergeCoverageStatus(statusPassed, "skipped"); got != statusPartial {
		t.Fatalf("non-passed observed should become partial, got %s", got)
	}
}

package main

import (
	"runtime"
	"testing"
	"time"
)

func assertFreshEvidence(t *testing.T, evidence evidenceFile) {
	t.Helper()
	if evidence.ExpiresAt == "" || evidence.FreshForSeconds != int64((24*time.Hour).Seconds()) {
		t.Fatalf("expiration evidence missing: %+v", evidence)
	}
	if evidence.Platform.OS != runtime.GOOS || evidence.Platform.Arch != runtime.GOARCH {
		t.Fatalf("platform=%+v, want %s/%s", evidence.Platform, runtime.GOOS, runtime.GOARCH)
	}
}

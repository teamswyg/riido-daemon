package openclaw

import (
	"strings"
	"testing"
)

func TestDetectOpenClawVersionSupportedFixture(t *testing.T) {
	res := detectWithFixture(t, "version_supported.txt", 0)

	if !res.Available {
		t.Fatalf("supported fixture must be Available: %+v", res)
	}
	if res.Version != "openclaw 2026.5.5" && !strings.Contains(res.Version, "2026.5.5") {
		t.Fatalf("Version: %q", res.Version)
	}
}

func TestDetectOpenClawVersionTooOldFixture(t *testing.T) {
	res := detectWithFixture(t, "version_too_old.txt", 0)

	if res.Available {
		t.Fatalf("too-old fixture must be unavailable: %+v", res)
	}
	if !strings.Contains(res.Reason, MinSupportedVersion) {
		t.Fatalf("Reason should mention minimum %s: %q", MinSupportedVersion, res.Reason)
	}
	if !strings.Contains(res.Version, "2026.5.4") {
		t.Fatalf("Version should still report what we observed: %q", res.Version)
	}
}

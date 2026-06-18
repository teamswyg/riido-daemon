package openclaw

import (
	"strings"
	"testing"
)

func TestDetectOpenClawVersionNonZeroDoesNotParseVersion(t *testing.T) {
	res := detectWithFixture(t, "version_node_error.txt", 1)

	if res.Available {
		t.Fatalf("non-zero --version MUST mark unavailable: %+v", res)
	}
	if res.Version != "" {
		t.Fatalf("Version must be empty on non-zero --version, got %q", res.Version)
	}
	if res.Reason == "" {
		t.Fatal("Reason must be set when --version fails")
	}
	if strings.Contains(res.Reason, "\n") {
		t.Fatalf("Reason must normalize newlines, got %q", res.Reason)
	}
	if len(res.Reason) > 400 {
		t.Fatalf("Reason too long: %d chars (%q)", len(res.Reason), res.Reason)
	}
	if !strings.Contains(res.Reason, "Node") {
		t.Fatalf("Reason should preserve human-actionable text, got %q", res.Reason)
	}
}

func TestDetectOpenClawVersionStacktrace(t *testing.T) {
	res := detectWithFixture(t, "version_stacktrace.txt", 1)

	if res.Available {
		t.Fatalf("non-zero exit must be unavailable: %+v", res)
	}
	if res.Version != "" {
		t.Fatalf("Version must be empty; got %q (looks like a Node version was lifted)", res.Version)
	}
	if strings.Contains(res.Version, "20.10") || strings.Contains(res.Version, "22.12") {
		t.Fatalf("Version field contains Node semver: %q", res.Version)
	}
}

func TestDetectOpenClawVersionNonZeroOverridesValidLookingOutput(t *testing.T) {
	res := detectWithFixture(t, "version_supported.txt", 2)

	if res.Available {
		t.Fatalf("non-zero exit must override valid-looking output: %+v", res)
	}
	if res.Version != "" {
		t.Fatalf("Version must be empty on non-zero exit: %q", res.Version)
	}
}

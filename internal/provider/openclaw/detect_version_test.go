package openclaw

import (
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// The load-bearing M-8 invariant: Node semver and embedded dependency
// numbers MUST NOT parse as OpenClaw versions.
func TestParseOpenClawVersionRejectsNodeSemver(t *testing.T) {
	cases := []string{
		"requires Node >=22.12.0",
		"node v20.10.0",
		"Node.js v20.10.0",
		"Detected: node 20.10.0 (exec: /usr/bin/node)",
		"package error: dep@22.12.0",
		"    at /path/22.12.0/file.js",
		"22.12.0",  // bare two-digit-prefix version
		"20.10.0",  // bare two-digit-prefix version
		"v20.10.0", // v-prefixed two-digit semver
		"v22.12.0", // v-prefixed two-digit semver
		"openclaw requires Node >=22.12.0",
	}
	for _, in := range cases {
		got, ok := parseVersion(in)
		if ok {
			t.Fatalf("parseVersion(%q) MUST reject Node-style semver; got %v", in, got)
		}
	}
}

// --- M-8 Detect fail-closed on non-zero exit ---

func TestDetectOpenClawVersionNonZeroDoesNotParseVersion(t *testing.T) {
	exe := writeShimFromFixture(t, "version_node_error.txt", 1)
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
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
	// Reason should preserve the human-actionable hint.
	if !strings.Contains(res.Reason, "Node") {
		t.Fatalf("Reason should preserve human-actionable text, got %q", res.Reason)
	}
}

// Stack-trace fixture: contains embedded "20.10.0" but exit was non-zero.
// Version must NOT be the embedded Node version.
func TestDetectOpenClawVersionStacktrace(t *testing.T) {
	exe := writeShimFromFixture(t, "version_stacktrace.txt", 1)
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
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

// Even if a non-zero shim emits a string that LOOKS like an OpenClaw
// version (e.g. `openclaw 2026.5.5`), a non-zero exit is treated as
// "no version" — exit code is authoritative.
func TestDetectOpenClawVersionNonZeroOverridesValidLookingOutput(t *testing.T) {
	exe := writeShimFromFixture(t, "version_supported.txt", 2)
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
	if res.Available {
		t.Fatalf("non-zero exit must override valid-looking output: %+v", res)
	}
	if res.Version != "" {
		t.Fatalf("Version must be empty on non-zero exit: %q", res.Version)
	}
}

// --- M-8 fixture-driven supported / too-old cases ---

func TestDetectOpenClawVersionSupportedFixture(t *testing.T) {
	exe := writeShimFromFixture(t, "version_supported.txt", 0)
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
	if !res.Available {
		t.Fatalf("supported fixture must be Available: %+v", res)
	}
	if res.Version != "openclaw 2026.5.5" && !strings.Contains(res.Version, "2026.5.5") {
		t.Fatalf("Version: %q", res.Version)
	}
}

func TestDetectOpenClawVersionTooOldFixture(t *testing.T) {
	exe := writeShimFromFixture(t, "version_too_old.txt", 0)
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
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

// --- Reason sanitizer unit tests ---

func TestSanitizeReasonHandlesEmpty(t *testing.T) {
	r := sanitizeReason("")
	if r == "" {
		t.Fatal("empty input must produce a non-empty fallback reason")
	}
}

func TestSanitizeReasonNormalizesNewlines(t *testing.T) {
	in := "line one\nline two\nline three"
	r := sanitizeReason(in)
	if strings.Contains(r, "\n") {
		t.Fatalf("newlines not normalized: %q", r)
	}
	if !strings.Contains(r, "line one") || !strings.Contains(r, "line three") {
		t.Fatalf("content lost: %q", r)
	}
}

func TestSanitizeReasonCapsLength(t *testing.T) {
	huge := strings.Repeat("a", 2000)
	r := sanitizeReason(huge)
	if len(r) > 400 {
		t.Fatalf("reason not capped: %d chars", len(r))
	}
}

// --- compareVersions sanity ---

func TestCompareVersions(t *testing.T) {
	if compareVersions([3]int{2026, 5, 5}, [3]int{2026, 5, 5}) != 0 {
		t.Fatal("equal")
	}
	if compareVersions([3]int{2026, 5, 4}, [3]int{2026, 5, 5}) != -1 {
		t.Fatal("less")
	}
	if compareVersions([3]int{2026, 6, 1}, [3]int{2026, 5, 31}) != 1 {
		t.Fatal("greater")
	}
}

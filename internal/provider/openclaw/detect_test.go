package openclaw

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// writeShim writes a script that echoes `version` and exits 0.
// Kept for the legacy test cases. New M-8 tests use writeShimFromFixture.
func writeShim(t *testing.T, version string) string {
	t.Helper()
	dir := t.TempDir()
	return writeShimInDir(t, dir, version)
}

func writeShimInDir(t *testing.T, dir, version string) string {
	t.Helper()
	path := filepath.Join(dir, "openclaw")
	script := "#!/bin/sh\necho '" + version + "'\nexit 0\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write shim: %v", err)
	}
	return path
}

// writeShimFromFixture writes a shim that cats the named testdata fixture
// to stdout (preserving multi-line content) and exits with exitCode.
// This is how M-8 simulates real-world `openclaw --version` output that
// can be a multi-line Node-dependency error with non-zero exit.
func writeShimFromFixture(t *testing.T, fixture string, exitCode int) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", fixture))
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixture, err)
	}
	dir := t.TempDir()
	contentPath := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(contentPath, body, 0o644); err != nil {
		t.Fatalf("write content: %v", err)
	}
	exePath := filepath.Join(dir, "openclaw")
	script := "#!/bin/sh\ncat " + contentPath + "\nexit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(exePath, []byte(script), 0o755); err != nil {
		t.Fatalf("write shim: %v", err)
	}
	return exePath
}

// --- Legacy detect tests (preserved) ---

func TestDetectMissingBinary(t *testing.T) {
	res, err := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: "/no/such/openclaw"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Available {
		t.Fatalf("Available: %+v", res)
	}
}

func TestDetectAcceptsAtMinimumVersion(t *testing.T) {
	exe := writeShim(t, MinSupportedVersion)
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
	if !res.Available {
		t.Fatalf("Available: %+v", res)
	}
}

func TestDetectRejectsOlderThanMinimum(t *testing.T) {
	exe := writeShim(t, "2026.4.30")
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
	if res.Available {
		t.Fatalf("expected gate to reject older version: %+v", res)
	}
	if !strings.Contains(res.Reason, MinSupportedVersion) {
		t.Fatalf("reason should mention minimum: %q", res.Reason)
	}
}

func TestDetectAcceptsNewerVersion(t *testing.T) {
	exe := writeShim(t, "v2026.12.31")
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
	if !res.Available {
		t.Fatalf("newer should pass: %+v", res)
	}
}

func TestDetectUnparseableVersion(t *testing.T) {
	exe := writeShim(t, "garbage-version")
	res, _ := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
	if res.Available {
		t.Fatalf("unparseable version must not be Available: %+v", res)
	}
}

func TestDetectScansPathCandidatesUntilSupportedVersion(t *testing.T) {
	oldDir := t.TempDir()
	newDir := t.TempDir()
	oldExe := writeShimInDir(t, oldDir, "OpenClaw 2026.3.24")
	newExe := writeShimInDir(t, newDir, "OpenClaw 2026.5.22")
	t.Setenv("PATH", oldDir+string(os.PathListSeparator)+newDir)

	res, err := Detect(context.Background(), agentbridge.DetectEnv{})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Available {
		t.Fatalf("later supported PATH candidate should be available: %+v", res)
	}
	if res.Executable != newExe {
		t.Fatalf("selected executable: got %q, want %q (old %q)", res.Executable, newExe, oldExe)
	}
	if !strings.Contains(res.Version, "2026.5.22") {
		t.Fatalf("Version should come from supported candidate, got %q", res.Version)
	}
	candidateCount, err := strconv.Atoi(res.Metadata["path_candidate_count"])
	if err != nil || candidateCount < 2 || res.Metadata["path_candidate_index"] != "2" {
		t.Fatalf("candidate metadata: %+v", res.Metadata)
	}
}

func TestDetectEnvOverridePinsOldVersionWithoutPathFallback(t *testing.T) {
	oldExe := writeShim(t, "OpenClaw 2026.3.24")
	newDir := t.TempDir()
	_ = writeShimInDir(t, newDir, "OpenClaw 2026.5.22")
	t.Setenv("PATH", newDir)

	res, err := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: oldExe},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Available {
		t.Fatalf("old explicit override must fail closed without PATH fallback: %+v", res)
	}
	if res.Executable != oldExe {
		t.Fatalf("override executable should be reported, got %q want %q", res.Executable, oldExe)
	}
	if !strings.Contains(res.Reason, MinSupportedVersion) {
		t.Fatalf("reason should mention minimum %s: %q", MinSupportedVersion, res.Reason)
	}
}

// --- M-8 parser hardness ---

func TestParseOpenClawVersionAcceptsDateStyle(t *testing.T) {
	cases := []struct {
		in   string
		want [3]int
	}{
		{"2026.5.5", [3]int{2026, 5, 5}},
		{"v2026.5.5", [3]int{2026, 5, 5}},
		{"openclaw 2026.5.5", [3]int{2026, 5, 5}},
		{"OpenClaw version 2026.05.05", [3]int{2026, 5, 5}},
		{"openclaw version 2026.12.31", [3]int{2026, 12, 31}},
	}
	for _, tc := range cases {
		got, ok := parseVersion(tc.in)
		if !ok || got != tc.want {
			t.Fatalf("parseVersion(%q): got %v ok=%v want %v", tc.in, got, ok, tc.want)
		}
	}
}

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

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

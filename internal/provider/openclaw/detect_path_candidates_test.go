package openclaw

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

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

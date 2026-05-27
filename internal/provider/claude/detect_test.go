package claude

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestDetectMissingBinary(t *testing.T) {
	res, err := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: "/no/such/path"},
	})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if res.Available {
		t.Fatalf("expected Available=false, got %+v", res)
	}
	if res.Reason == "" {
		t.Fatal("expected non-empty Reason")
	}
}

func TestDetectFakeBinaryReportsVersion(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "claude")
	script := "#!/bin/sh\necho '1.5.7 (anthropic-claude-code)'\nexit 0\n"
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: fake},
	})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !res.Available {
		t.Fatalf("Available: %+v", res)
	}
	if res.Executable != fake {
		t.Fatalf("Executable: %q", res.Executable)
	}
	if res.Version == "" {
		t.Fatalf("Version empty: %+v", res)
	}
	if !res.SupportsStreaming || !res.SupportsResume {
		t.Fatalf("capability flags wrong: %+v", res)
	}
}

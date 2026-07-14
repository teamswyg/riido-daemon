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

func TestDetectOverrideReportsVersion(t *testing.T) {
	exe := writeClaudeProbe(t, true)

	res, err := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: exe},
	})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !res.Available {
		t.Fatalf("Available: %+v", res)
	}
	if res.Executable != exe {
		t.Fatalf("Executable: %q", res.Executable)
	}
	if res.Version == "" {
		t.Fatalf("Version empty: %+v", res)
	}
	if !res.SupportsStreaming || !res.SupportsResume {
		t.Fatalf("capability flags wrong: %+v", res)
	}
}

func TestDetectLoggedOutIsUnavailable(t *testing.T) {
	res, err := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: writeClaudeProbe(t, false)},
	})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if res.Available || res.Reason != claudeAuthRecoveryMessage {
		t.Fatalf("result: %+v", res)
	}
}

func writeClaudeProbe(t *testing.T, loggedIn bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "claude")
	status := "false"
	if loggedIn {
		status = "true"
	}
	script := "#!/bin/sh\nif [ \"$1 $2\" = \"auth status\" ]; then echo '{\"loggedIn\":" + status + "}'; else echo '2.1.202'; fi\n"
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

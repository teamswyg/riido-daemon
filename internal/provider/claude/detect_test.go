package claude

import (
	"context"
	"os/exec"
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
	echo, err := exec.LookPath("echo")
	if err != nil {
		t.Fatal(err)
	}

	res, err := Detect(context.Background(), agentbridge.DetectEnv{
		EnvOverride: map[string]string{EnvOverride: echo},
	})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !res.Available {
		t.Fatalf("Available: %+v", res)
	}
	if res.Executable != echo {
		t.Fatalf("Executable: %q", res.Executable)
	}
	if res.Version == "" {
		t.Fatalf("Version empty: %+v", res)
	}
	if !res.SupportsStreaming || !res.SupportsResume {
		t.Fatalf("capability flags wrong: %+v", res)
	}
}

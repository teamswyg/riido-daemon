package main

import (
	"context"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"path/filepath"
	"testing"
)

func TestDesktopProviderPathPinsReachDetection(t *testing.T) {
	for _, entry := range []struct {
		key     string
		adapter agentbridge.Adapter
	}{{"RIIDO_CLAUDE_PATH", bridgeClaudeAdapter{}}, {"RIIDO_CODEX_PATH", bridgeCodexAdapter{}}, {"RIIDO_OPENCLAW_PATH", bridgeOpenClawAdapter{}}, {"RIIDO_CURSOR_PATH", bridgeCursorAdapter{}}} {
		t.Run(entry.key, func(t *testing.T) {
			t.Setenv(entry.key, filepath.Join(t.TempDir(), "missing-provider"))
			result, err := entry.adapter.Detect(context.Background(), agentbridge.DetectEnv{})
			if err != nil {
				t.Fatal(err)
			}
			if result.Available {
				t.Fatalf("ignored explicit host pin and selected %s", result.Executable)
			}
		})
	}
}

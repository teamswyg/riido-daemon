package main

import (
	"os"
	"path/filepath"
)

func desktopDaemonStopEvidencePath() string {
	return filepath.Join(
		desktopAppSupportDir(),
		"riido-desktop",
		"ai-agent-daemon",
		"daemon-stop-events.jsonl",
	)
}

func desktopAgentdSocketPath() string {
	return filepath.Join(desktopAppSupportDir(), "riido", "agentd.sock")
}

func desktopAppSupportDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, "Library", "Application Support")
}

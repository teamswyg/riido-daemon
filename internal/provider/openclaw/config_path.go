package openclaw

import (
	"os"
	"path/filepath"
	"strings"
)

func sourceConfigPath(env map[string]string) (string, bool) {
	if path := strings.TrimSpace(env[openClawConfigPathEnv]); path != "" {
		return resolveUserPath(path), true
	}
	if stateDir := strings.TrimSpace(env[openClawStateDirEnv]); stateDir != "" {
		return filepath.Join(resolveUserPath(stateDir), "openclaw.json"), true
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "", false
	}
	return filepath.Join(home, ".openclaw", "openclaw.json"), true
}

func resolveUserPath(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}

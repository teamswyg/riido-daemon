package workdir

import (
	"path/filepath"
	"strings"
)

func removeHookSettingsFiles(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == ".claude/settings.json" {
			continue
		}
		out = append(out, path)
	}
	return out
}

func removeConfigHomeSettingsFiles(paths []string, configHomeDir string) []string {
	configHomeDir = filepath.ToSlash(strings.TrimSpace(configHomeDir))
	if configHomeDir == "" {
		return append([]string(nil), paths...)
	}
	prefix := strings.TrimSuffix(configHomeDir, "/") + "/"
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if strings.HasPrefix(filepath.ToSlash(strings.TrimSpace(path)), prefix) {
			continue
		}
		out = append(out, path)
	}
	return out
}

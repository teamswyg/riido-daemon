package detectutil

import (
	"os"
	"path/filepath"
	"strings"
)

func windowsWellKnownDirs() []string {
	var dirs []string
	if appData := strings.TrimSpace(os.Getenv("APPDATA")); appData != "" {
		dirs = append(dirs, filepath.Join(appData, "npm"))
	}
	home, err := userHomeDir()
	if err == nil && strings.TrimSpace(home) != "" {
		dirs = append(dirs, windowsHomeDirs(home)...)
	}
	if localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); localAppData != "" {
		dirs = append(dirs, filepath.Join(localAppData, "Programs"))
	}
	return dirs
}

func windowsHomeDirs(home string) []string {
	return []string{
		filepath.Join(home, ".cargo", "bin"),
		filepath.Join(home, ".bun", "bin"),
		filepath.Join(home, "go", "bin"),
		filepath.Join(home, ".cursor", "bin"),
		filepath.Join(home, ".claude", "bin"),
	}
}

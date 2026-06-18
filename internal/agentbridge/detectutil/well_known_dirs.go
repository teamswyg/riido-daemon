package detectutil

import (
	"path/filepath"
	"runtime"
	"strings"
)

func wellKnownInstallDirs() []string {
	if runtime.GOOS == "windows" {
		return windowsWellKnownDirs()
	}
	dirs := unixWellKnownDirs()
	home, err := userHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return dirs
	}
	return append(dirs, userInstallDirs(home)...)
}

func unixWellKnownDirs() []string {
	return []string{
		"/opt/homebrew/bin", "/opt/homebrew/sbin",
		"/usr/local/bin", "/usr/local/sbin",
		"/usr/bin", "/bin", "/usr/sbin", "/sbin",
	}
}

func userInstallDirs(home string) []string {
	dirs := []string{
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, "bin"),
		filepath.Join(home, ".npm-global", "bin"),
		filepath.Join(home, ".cargo", "bin"),
		filepath.Join(home, ".bun", "bin"),
		filepath.Join(home, ".deno", "bin"),
		filepath.Join(home, "go", "bin"),
		filepath.Join(home, ".volta", "bin"),
		filepath.Join(home, ".asdf", "shims"),
		filepath.Join(home, ".cursor", "bin"),
		filepath.Join(home, ".claude", "bin"),
	}
	return append(dirs, nodeVersionManagerBins(home)...)
}

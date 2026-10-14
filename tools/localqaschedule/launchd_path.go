package main

import (
	"os"
	"path/filepath"
	"strings"
)

func launchdPath() string {
	parts := []string{}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		parts = append(parts, filepath.Join(home, ".local", "bin"))
		parts = append(parts, filepath.Join(home, "bin"))
	}
	parts = append(parts,
		"/opt/homebrew/bin",
		"/usr/local/bin",
		"/usr/local/go/bin",
		"/usr/bin",
		"/bin",
		"/usr/sbin",
		"/sbin",
	)
	return strings.Join(parts, ":")
}

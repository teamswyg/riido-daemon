package main

import (
	"path/filepath"
	"strings"
)

func resolvePath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.FromSlash(path))
}

func executableRef(provider provider, overridePath string, found bool) string {
	if !found {
		return ""
	}
	if overridePath != "" {
		return "$" + provider.OverrideEnv
	}
	return provider.DefaultExecutable
}

func compactOutput(out string) string {
	out = strings.TrimSpace(out)
	if len(out) <= 600 {
		return out
	}
	return out[len(out)-600:]
}

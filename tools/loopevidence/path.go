package main

import (
	"path/filepath"
	"strings"
)

const defaultManifest = "docs/30-architecture/loop-engineering.riido.json"

func resolvePath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.FromSlash(path))
}

func isCommandRef(ref string) bool {
	return strings.Contains(ref, " ") || strings.Contains(ref, "\t")
}

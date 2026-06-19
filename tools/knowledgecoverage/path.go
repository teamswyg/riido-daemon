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

func slashPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

func siblingManifest(path string) string {
	return strings.TrimSuffix(path, ".md") + ".riido.json"
}

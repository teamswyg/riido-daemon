package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func findRepoRoot(start string) (string, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(abs, "go.mod")); err == nil {
			return abs, nil
		}
		next := filepath.Dir(abs)
		if next == abs {
			return "", fmt.Errorf("go.mod not found from %s", start)
		}
		abs = next
	}
}

func repoPath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

func slashPath(path string) string {
	return strings.TrimPrefix(filepath.ToSlash(path), "./")
}

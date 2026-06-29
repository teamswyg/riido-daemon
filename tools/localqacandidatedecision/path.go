package main

import (
	"os"
	"path/filepath"
)

func findRepoRoot(start string) (string, error) {
	if start == "" {
		start = "."
	}
	root, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if localFileExists(filepath.Join(root, "go.mod")) {
			return root, nil
		}
		parent := filepath.Dir(root)
		if parent == root {
			return root, nil
		}
		root = parent
	}
}

func repoPath(root, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(root, filepath.FromSlash(rel))
}

func localFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

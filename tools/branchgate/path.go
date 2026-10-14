package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

const defaultManifest = "docs/30-architecture/riido-work-branch-gate.riido.json"

func cleanRepoPath(repo, rel string) (string, error) {
	if filepath.IsAbs(rel) || strings.Contains(rel, "..") {
		return "", fmt.Errorf("unsafe repo path: %s", rel)
	}
	return filepath.Join(repo, filepath.Clean(rel)), nil
}

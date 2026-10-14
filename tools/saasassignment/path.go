package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

const defaultManifest = "docs/20-domain/runtime-scheduling/saas-assignment-source.riido.json"

func cleanRepoPath(repo, rel string) (string, error) {
	if filepath.IsAbs(rel) || strings.Contains(rel, "..") {
		return "", fmt.Errorf("unsafe repo path: %s", rel)
	}
	return filepath.Join(repo, filepath.Clean(rel)), nil
}

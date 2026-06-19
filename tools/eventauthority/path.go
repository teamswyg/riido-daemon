package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

const defaultManifest = "docs/20-domain/provider-runtime/adapter-acl/event-ingestor-contract.riido.json"

func cleanRepoPath(repo, p string) (string, error) {
	if filepath.IsAbs(p) {
		return "", fmt.Errorf("path must be repo relative: %s", p)
	}
	clean := filepath.Clean(p)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes repo: %s", p)
	}
	return filepath.Join(repo, clean), nil
}

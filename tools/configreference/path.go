package main

import (
	"path/filepath"
	"strings"
)

const defaultManifest = "docs/30-architecture/config-reference.riido.json"

func repoPath(repo, rel string) string {
	return filepath.Join(repo, filepath.Clean(rel))
}

func normalizeOptions(opts options) options {
	if strings.TrimSpace(opts.Repo) == "" {
		opts.Repo = "."
	}
	if strings.TrimSpace(opts.Manifest) == "" {
		opts.Manifest = defaultManifest
	}
	return opts
}

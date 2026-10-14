package main

import "path/filepath"

const defaultManifest = "docs/20-domain/provider-runtime/public-migration-status.riido.json"

func repoPath(repo, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(repo, filepath.FromSlash(rel))
}

package main

import "path/filepath"

const defaultManifest = "docs/30-architecture/config-reference/executable-search-path.riido.json"

func repoPath(repo, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(repo, rel)
}

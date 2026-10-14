package main

import "path/filepath"

const defaultManifest = "docs/50-roadmap/open-questions.riido.json"

func repoPath(repo, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(repo, filepath.FromSlash(rel))
}

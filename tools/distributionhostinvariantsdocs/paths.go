package main

import "path/filepath"

const defaultManifest = "docs/20-domain/distribution-host-integration/invariants.riido.json"

func repoPath(repo, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(repo, filepath.FromSlash(rel))
}

func fragmentPath(repo, manifestPath, rel string) string {
	base := filepath.Dir(repoPath(repo, manifestPath))
	return filepath.Join(base, filepath.FromSlash(rel))
}

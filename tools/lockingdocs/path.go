package main

import "path/filepath"

const defaultManifest = "docs/20-domain/locking.riido.json"

func repoPath(repo, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(repo, filepath.FromSlash(rel))
}

func fragmentPath(repo, manifestPath, rel string) string {
	return filepath.Join(filepath.Dir(repoPath(repo, manifestPath)), filepath.FromSlash(rel))
}

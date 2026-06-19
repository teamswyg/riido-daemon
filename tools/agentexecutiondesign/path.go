package main

import "path/filepath"

func repoPath(repo, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(repo, filepath.FromSlash(rel))
}

func manifestBase(path string) string {
	return filepath.Dir(filepath.ToSlash(path))
}

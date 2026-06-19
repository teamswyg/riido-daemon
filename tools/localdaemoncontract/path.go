package main

import "path/filepath"

const defaultManifest = "docs/20-domain/runtime-scheduling/invariants/local-daemon-contract.riido.json"

func repoPath(repo, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(repo, rel)
}

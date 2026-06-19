package main

import "path/filepath"

const defaultManifest = "docs/20-domain/distribution-host-integration/store-channel-policy/server-facing-metadata.riido.json"

func repoPath(repo, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(repo, rel)
}

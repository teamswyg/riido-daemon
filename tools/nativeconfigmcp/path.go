package main

import "path/filepath"

const defaultManifest = "docs/20-domain/security/enforcement-locations/native-config-and-mcp.riido.json"

func repoPath(repo, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(repo, rel)
}

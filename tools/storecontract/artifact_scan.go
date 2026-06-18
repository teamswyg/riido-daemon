package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func scanArtifactRoots(repoRoot string, roots, providerNames []string) []string {
	var problems []string
	for _, root := range roots {
		path := resolvePath(repoRoot, root)
		info, err := os.Stat(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("store artifact root missing: %s", root))
			continue
		}
		if !info.IsDir() {
			problems = append(problems, fmt.Sprintf("store artifact root is not a directory: %s", root))
			continue
		}
		walkErr := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				problems = append(problems, fmt.Sprintf("scan %s: %v", path, err))
				return nil
			}
			if entry.IsDir() {
				return nil
			}
			if matchesProviderBinary(entry.Name(), providerNames) {
				problems = append(problems, fmt.Sprintf("provider CLI appears bundled in store artifact root: %s", path))
			}
			if hasHardcodedUserPath(path) {
				problems = append(problems, fmt.Sprintf("store artifact contains hardcoded user path: %s", path))
			}
			return nil
		})
		if walkErr != nil {
			problems = append(problems, fmt.Sprintf("scan root %s: %v", root, walkErr))
		}
	}
	return problems
}

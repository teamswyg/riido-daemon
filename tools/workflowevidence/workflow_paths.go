package main

import (
	"path/filepath"
	"sort"
)

func workflowPaths(root, workflowRoot string) ([]string, error) {
	var paths []string
	for _, pattern := range []string{"*.yml", "*.yaml"} {
		found, err := filepath.Glob(repoPath(root, filepath.Join(workflowRoot, pattern)))
		if err != nil {
			return nil, err
		}
		paths = append(paths, found...)
	}
	sort.Strings(paths)
	return uniqueStrings(paths), nil
}

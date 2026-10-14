package main

import (
	"os"
	"path/filepath"
	"strings"
)

func closedLoopMaturityMeta(root string) map[string]any {
	return map[string]any{
		"tool_dir_count":             toolDirCount(filepath.Join(root, "tools")),
		"workflow_count":             globCount(root, ".github/workflows/*.yml"),
		"verifier_count":             verifierCount(root),
		"generated_artifact_count":   suffixCount(filepath.Join(root, "tools"), ".generated.json"),
		"loop_file_count":            globCount(root, "loopregistry/*.json"),
		"entrypoint_candidate_count": mainCount(filepath.Join(root, "tools")),
	}
}

func toolDirCount(root string) int {
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			count++
		}
	}
	return count
}

func verifierCount(root string) int {
	count := 0
	_ = filepath.WalkDir(filepath.Join(root, "tools"), func(file string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(file, ".go") {
			return nil
		}
		base := filepath.Base(file)
		if strings.Contains(base, "verify") || strings.HasSuffix(base, "_test.go") {
			count++
		}
		return nil
	})
	return count
}

func globCount(root, pattern string) int {
	files, err := filepath.Glob(filepath.Join(root, pattern))
	if err != nil {
		return 0
	}
	return len(files)
}

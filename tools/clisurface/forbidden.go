package main

import (
	"os"
	"path/filepath"
	"strings"
)

func checkForbidden(repo string, manifest Manifest) []CheckResult {
	results := make([]CheckResult, 0, len(manifest.ForbiddenSourceTokens))
	for _, check := range manifest.ForbiddenSourceTokens {
		results = append(results, checkForbiddenToken(repo, check))
	}
	return results
}

func checkForbiddenToken(repo string, check ForbiddenSourceToken) CheckResult {
	result := CheckResult{Name: check.Name, File: check.Root, Pass: true}
	root := repoPath(repo, check.Root)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, check.Suffix) {
			return err
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if strings.Contains(string(data), check.Contains) {
			result.Pass = false
			result.File = path
			result.Detail = "forbidden token found"
		}
		return nil
	})
	if err != nil {
		result.Pass = false
		result.Detail = err.Error()
	}
	return result
}

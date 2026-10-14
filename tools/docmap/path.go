package main

import (
	"path/filepath"
	"strings"
)

func resolvePath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.FromSlash(path))
}

func linkFrom(docPath, target string) string {
	fromDir := filepath.Dir(filepath.FromSlash(docPath))
	rel, err := filepath.Rel(fromDir, filepath.FromSlash(target))
	if err != nil {
		return target
	}
	return filepath.ToSlash(rel)
}

func linkLabel(path string) string {
	return strings.TrimPrefix(path, "docs/")
}

func joinProblems(problems []string) string {
	var out strings.Builder
	for _, problem := range problems {
		out.WriteString("- ")
		out.WriteString(problem)
		out.WriteByte('\n')
	}
	return out.String()
}

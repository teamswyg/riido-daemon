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

func joinProblems(problems []string) string {
	var out strings.Builder
	for _, problem := range problems {
		out.WriteString("- ")
		out.WriteString(problem)
		out.WriteByte('\n')
	}
	return out.String()
}

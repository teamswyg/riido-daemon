package main

import (
	"os"
	"strings"
)

func readChangedFiles(root, rel string) ([]string, error) {
	body, err := os.ReadFile(repoPath(root, rel))
	if err != nil {
		return nil, err
	}
	var out []string
	for line := range strings.SplitSeq(string(body), "\n") {
		if trimmed := slash(line); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out, nil
}

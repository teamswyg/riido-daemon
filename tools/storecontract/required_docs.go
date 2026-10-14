package main

import (
	"fmt"
	"os"
)

func validateRequiredDocs(repoRoot string, docs []string) []string {
	var problems []string
	for _, doc := range docs {
		path := resolvePath(repoRoot, doc)
		info, err := os.Stat(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("required doc missing: %s", doc))
			continue
		}
		if info.IsDir() {
			problems = append(problems, fmt.Sprintf("required doc is a directory: %s", doc))
		}
	}
	return problems
}

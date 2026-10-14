package main

import (
	"fmt"
	"os"
)

func mustExist(repo, rel string) []string {
	if rel == "" {
		return nil
	}
	if _, err := os.Stat(repoPath(repo, rel)); err != nil {
		return []string{fmt.Sprintf("missing artifact %q", rel)}
	}
	return nil
}

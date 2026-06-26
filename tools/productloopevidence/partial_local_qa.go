package main

import "os"

func localQARunPresent(root, rel string) bool {
	if rel == "" {
		return false
	}
	_, err := os.Stat(repoPath(root, rel))
	return err == nil
}

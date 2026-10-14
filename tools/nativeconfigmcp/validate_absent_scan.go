package main

import (
	"os"
	"path/filepath"
	"strings"
)

func scopeContains(root, token string) (bool, error) {
	info, err := os.Stat(root)
	if err != nil {
		return false, err
	}
	if !info.IsDir() {
		return fileContains(root, token)
	}
	hit := false
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || hit || entry.IsDir() || filepath.Ext(path) != ".go" {
			return err
		}
		found, readErr := fileContains(path, token)
		if readErr != nil {
			return readErr
		}
		hit = found
		return nil
	})
	return hit, err
}

func fileContains(path, token string) (bool, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(body), token), nil
}

package main

import (
	"os"
	"path/filepath"
	"strings"
)

func suffixCount(root, suffix string) int {
	count := 0
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, suffix) {
			count++
		}
		return nil
	})
	return count
}

func mainCount(root string) int {
	count := 0
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Base(path) == "main.go" {
			count++
		}
		return nil
	})
	return count
}

package main

import (
	"os"
	"path/filepath"
)

func manifestLoopReferencePath(root, ownerPath, source string) (string, bool) {
	if filepath.IsAbs(source) {
		return manifestSourcePath(root, source)
	}
	if target, ok := manifestSiblingSourcePath(root, ownerPath, source); ok && fileExists(target) {
		return target, true
	}
	return manifestSourcePath(root, source)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

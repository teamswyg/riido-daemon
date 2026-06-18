//go:build !windows

package main

import (
	"os"
	"path/filepath"
)

func daemonCommandBinaryLooksLikeSelf(argv0 string) bool {
	current, err := os.Executable()
	if err != nil {
		return false
	}
	binaryName := filepath.Base(argv0)
	currentName := filepath.Base(current)
	return binaryName == currentName || binaryName == "riido"
}

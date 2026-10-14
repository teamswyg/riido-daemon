package main

import (
	"os"
	"strings"
)

const featureUICapturePath = ".riido-local/screenshots/contract-lab/feature-ui-manual-pass.png"

func localFileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func captureCoveredByUploadDir(dir string) bool {
	return strings.HasPrefix(featureUICapturePath, strings.TrimRight(dir, "/")+"/")
}

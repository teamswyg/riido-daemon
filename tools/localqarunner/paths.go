package main

import (
	"os"
	"path/filepath"
)

func runEvidenceAbs(root string, cfg config) string {
	return outputPath(root, *cfg.runEvidence)
}

func outputPath(root, path string) string {
	clean := filepath.FromSlash(path)
	if filepath.IsAbs(clean) {
		return clean
	}
	return filepath.Join(root, clean)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func writeGeneratedFile(path string, rendered []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("riidogen: create output dir: %w", err)
	}
	if err := os.WriteFile(path, rendered, 0o644); err != nil {
		return fmt.Errorf("riidogen: write output: %w", err)
	}
	return nil
}

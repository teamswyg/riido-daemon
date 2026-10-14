package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func outputPath(root, path string) string {
	clean := filepath.FromSlash(path)
	if filepath.IsAbs(clean) {
		return clean
	}
	return filepath.Join(root, clean)
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create evidence dir: %w", err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal evidence: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write evidence: %w", err)
	}
	return nil
}

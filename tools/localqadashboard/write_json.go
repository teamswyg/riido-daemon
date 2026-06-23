package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeJSON(path string, value any) error {
	if path == "" {
		return nil
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create json dir: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write json: %w", err)
	}
	return nil
}

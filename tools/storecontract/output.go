package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeJSON(path string, value checkResult) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode check output: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write check output: %w", err)
	}
	return nil
}

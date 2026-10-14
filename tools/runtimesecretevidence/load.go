package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadManifest(path string) (manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	var out manifest
	if err := json.Unmarshal(data, &out); err != nil {
		return manifest{}, fmt.Errorf("parse manifest: %w", err)
	}
	return out, nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write evidence: %w", err)
	}
	return nil
}

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

func writeText(path, text string) error {
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return fmt.Errorf("write generated doc: %w", err)
	}
	return nil
}

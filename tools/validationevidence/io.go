package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadManifest(repo, path string) (Manifest, error) {
	data, err := os.ReadFile(repoPath(repo, path))
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	var out Manifest
	if err := json.Unmarshal(data, &out); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest: %w", err)
	}
	return out, nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func writeText(path, text string) error {
	return os.WriteFile(path, []byte(text), 0o644)
}

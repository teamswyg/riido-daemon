package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadManifest(path string) (manifest, error) {
	var m manifest
	body, err := os.ReadFile(path)
	if err != nil {
		return m, fmt.Errorf("read manifest: %w", err)
	}
	if err := json.Unmarshal(body, &m); err != nil {
		return m, fmt.Errorf("decode manifest: %w", err)
	}
	return m, validateManifest(m)
}

func readEvidence(path string) (map[string]any, error) {
	var out map[string]any
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read evidence %s: %w", path, err)
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode evidence %s: %w", path, err)
	}
	return out, nil
}

func writeJSON(path string, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	body = append(body, '\n')
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

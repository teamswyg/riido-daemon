package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func loadManifest(repo, rel string) (manifest, error) {
	data, err := os.ReadFile(repoPath(repo, rel))
	if err != nil {
		return manifest{}, err
	}
	var m manifest
	return m, json.Unmarshal(data, &m)
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func writeText(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func loadManifest(repo, rel string) (manifest, error) {
	var m manifest
	if err := readJSON(repoPath(repo, rel), &m); err != nil {
		return m, err
	}
	if err := loadFragments(repo, rel, &m); err != nil {
		return m, err
	}
	return m, nil
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func writeText(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeText(path, string(append(data, '\n')))
}

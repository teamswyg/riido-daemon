package main

import (
	"encoding/json"
	"os"
)

func loadManifest(repo, path string) (Manifest, error) {
	full, err := cleanRepoPath(repo, path)
	if err != nil {
		return Manifest{}, err
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

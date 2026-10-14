package main

import (
	"encoding/json"
	"os"
)

func loadManifest(repo, rel string) (Manifest, error) {
	data, err := os.ReadFile(repoPath(repo, rel))
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
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func writeText(path, value string) error {
	return os.WriteFile(path, []byte(value), 0o644)
}

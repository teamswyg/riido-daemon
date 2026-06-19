package main

import (
	"encoding/json"
	"os"
)

func loadManifest(repo, rel string) (Manifest, error) {
	body, err := os.ReadFile(repoPath(repo, rel))
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func writeJSON(path string, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return os.WriteFile(path, body, 0o644)
}

func writeText(path, value string) error {
	return os.WriteFile(path, []byte(value), 0o644)
}

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
	return manifest, json.Unmarshal(data, &manifest)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func writeText(path, body string) error {
	return os.WriteFile(path, []byte(body), 0o644)
}

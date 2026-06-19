package main

import (
	"encoding/json"
	"os"
)

func loadManifest(repo, rel string) (Manifest, error) {
	path, err := cleanRepoPath(repo, rel)
	if err != nil {
		return Manifest{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	return manifest, json.Unmarshal(data, &manifest)
}

func readFile(repo, rel string) (string, error) {
	path, err := cleanRepoPath(repo, rel)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	return string(data), err
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

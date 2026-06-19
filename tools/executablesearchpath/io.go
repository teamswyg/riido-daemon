package main

import (
	"encoding/json"
	"os"
)

func loadManifest(repo, rel string) (Manifest, error) {
	var out Manifest
	raw, err := os.ReadFile(repoPath(repo, rel))
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(raw, &out)
	return out, err
}

func writeText(path, body string) error {
	return os.WriteFile(path, []byte(body), 0o644)
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

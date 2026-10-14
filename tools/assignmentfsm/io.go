package main

import (
	"bytes"
	"encoding/json"
	"os"
)

func loadManifest(repo, rel string) (Manifest, error) {
	var manifest Manifest
	body, err := os.ReadFile(repoPath(repo, rel))
	if err != nil {
		return manifest, err
	}
	if err := json.Unmarshal(body, &manifest); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func writeText(path, body string) error {
	return os.WriteFile(path, []byte(body), 0o644)
}

func writeJSON(path string, value any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

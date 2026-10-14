package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func loadManifest(repo, path string) (Manifest, error) {
	full := path
	if !filepath.IsAbs(full) {
		full = filepath.Join(repo, path)
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

func writeEvidence(path string, evidence Evidence) error {
	if path == "" {
		return nil
	}
	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func pathExists(repo, path string) bool {
	_, err := os.Stat(filepath.Join(repo, filepath.FromSlash(path)))
	return err == nil
}

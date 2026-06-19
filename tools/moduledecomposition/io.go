package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func loadManifest(repo, rel string) (manifest, error) {
	var m manifest
	if err := readJSON(repoPath(repo, rel), &m); err != nil {
		return manifest{}, err
	}
	base := filepath.Dir(rel)
	return m, loadFragments(repo, base, &m)
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func loadFragments(repo, base string, m *manifest) error {
	if err := loadFragment(repo, base, m.PackageRolesFile, &m.PackageRoles); err != nil {
		return err
	}
	if err := loadFragment(repo, base, m.ImportRulesFile, &m.ImportRules); err != nil {
		return err
	}
	return loadFragment(repo, base, m.PortsFile, &m.Ports)
}

func loadFragment(repo, base, rel string, value any) error {
	if rel == "" {
		return nil
	}
	return readJSON(repoPath(repo, filepath.Join(base, rel)), value)
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

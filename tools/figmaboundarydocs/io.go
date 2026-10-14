package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func loadManifest(repo, rel string) (manifest, error) {
	data, err := os.ReadFile(repoPath(repo, rel))
	if err != nil {
		return manifest{}, err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return manifest{}, err
	}
	entries, err := loadEntryFiles(repoPath(repo, rel), m.EntryFiles)
	if err != nil {
		return manifest{}, err
	}
	m.Entries = entries
	return m, nil
}

func loadEntryFiles(manifestPath string, files []string) ([]boundaryEntry, error) {
	var entries []boundaryEntry
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(filepath.Dir(manifestPath), filepath.FromSlash(file)))
		if err != nil {
			return nil, err
		}
		var loaded []boundaryEntry
		if err := json.Unmarshal(data, &loaded); err != nil {
			return nil, err
		}
		entries = append(entries, loaded...)
	}
	return entries, nil
}

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

func loadManifest(repo, rel string) (manifest, error) {
	data, err := os.ReadFile(repoPath(repo, rel))
	if err != nil {
		return manifest{}, err
	}
	var m manifest
	if err := decodeJSON(data, &m); err != nil {
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
		if err := decodeJSON(data, &loaded); err != nil {
			return nil, err
		}
		entries = append(entries, loaded...)
	}
	return entries, nil
}

func decodeJSON(data []byte, target any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return err
	}
	return requireEOF(dec)
}

func requireEOF(dec *json.Decoder) error {
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}

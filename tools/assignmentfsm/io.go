package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
)

func loadManifest(repo, rel string) (Manifest, error) {
	var manifest Manifest
	body, err := os.ReadFile(repoPath(repo, rel))
	if err != nil {
		return manifest, err
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&manifest); err != nil {
		return manifest, err
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return manifest, errors.New("trailing JSON value")
		}
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

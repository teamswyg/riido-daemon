package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func loadManifest(repo, path string) (Manifest, error) {
	data, err := os.ReadFile(repoPath(repo, path))
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	var out Manifest
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest: %w", err)
	}
	if err := requireJSONEOF(dec); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest: %w", err)
	}
	return out, nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("prepare evidence dir: %w", err)
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func writeText(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("prepare text dir: %w", err)
	}
	return os.WriteFile(path, []byte(text), 0o644)
}

func requireJSONEOF(dec *json.Decoder) error {
	var extra any
	err := dec.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("trailing JSON value")
	}
	return err
}

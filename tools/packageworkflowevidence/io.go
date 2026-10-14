package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func loadManifest(path string) (manifest, error) {
	m, err := readManifest(path)
	if err != nil {
		return m, err
	}
	if err := loadWorkflowSources(filepath.Dir(path), &m); err != nil {
		return m, err
	}
	return m, validateManifest(m)
}

func readManifest(path string) (manifest, error) {
	var m manifest
	data, err := os.ReadFile(path)
	if err != nil {
		return m, fmt.Errorf("read manifest: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return m, fmt.Errorf("decode manifest: %w", err)
	}
	return m, nil
}

func loadWorkflowSources(base string, m *manifest) error {
	for _, source := range m.WorkflowSources {
		part, err := readManifest(filepath.Join(base, source))
		if err != nil {
			return err
		}
		if part.SchemaVersion != manifestSchema || len(part.Workflows) == 0 {
			return fmt.Errorf("invalid workflow source %s", source)
		}
		m.Workflows = append(m.Workflows, part.Workflows...)
	}
	return nil
}

func writeJSON(path string, value evidence) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create evidence dir: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write evidence: %w", err)
	}
	return nil
}

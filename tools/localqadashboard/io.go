package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func loadProviderEvidence(path string) (providerEvidenceFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return providerEvidenceFile{}, fmt.Errorf("read provider evidence: %w", err)
	}
	var out providerEvidenceFile
	if err := json.Unmarshal(data, &out); err != nil {
		return providerEvidenceFile{}, fmt.Errorf("parse provider evidence: %w", err)
	}
	return out, nil
}

func loadCoverageManifest(path string) (coverageManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return coverageManifest{}, fmt.Errorf("read coverage manifest: %w", err)
	}
	var out coverageManifest
	if err := json.Unmarshal(data, &out); err != nil {
		return coverageManifest{}, fmt.Errorf("parse coverage manifest: %w", err)
	}
	return out, nil
}

func loadExternalEvidence(path string) (externalEvidenceFile, error) {
	if path == "" {
		return externalEvidenceFile{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return externalEvidenceFile{}, nil
		}
		return externalEvidenceFile{}, fmt.Errorf("read external evidence: %w", err)
	}
	var out externalEvidenceFile
	if err := json.Unmarshal(data, &out); err != nil {
		return externalEvidenceFile{}, fmt.Errorf("parse external evidence: %w", err)
	}
	return out, nil
}

func writeText(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create dashboard dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return fmt.Errorf("write dashboard: %w", err)
	}
	return nil
}

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeEvidence(path string, evidence evidenceFile) error {
	data, err := json.MarshalIndent(evidence, "", "  ")
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

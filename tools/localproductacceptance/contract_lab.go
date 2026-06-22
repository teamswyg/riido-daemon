package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

func writeContractLab(path string, evidence evidenceFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create contract lab dir: %w", err)
	}
	data, err := json.Marshal(evidence)
	if err != nil {
		return fmt.Errorf("encode contract lab evidence: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create contract lab: %w", err)
	}
	defer file.Close()
	return contractLabTemplate.Execute(file, map[string]any{
		"Evidence": template.JS(data),
	})
}

func contractUIScenario(path string) scenario {
	return scenario{
		ID:       "contract.ui.lab",
		Status:   statusPassed,
		Endpoint: path,
		Observed: map[string]any{
			"react_lab_html": path,
			"purpose":        "frontend API usage handoff",
		},
	}
}

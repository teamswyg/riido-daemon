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
	i18n, err := qaI18NJSON()
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create contract lab: %w", err)
	}
	defer file.Close()
	return contractLabTemplate.Execute(file, map[string]any{
		"Evidence": template.JS(data),
		"I18N":     template.JS(i18n),
	})
}

func contractUIScenario(path, manualOut string) scenario {
	return scenario{
		ID:       "contract.ui.lab",
		Status:   statusPassed,
		Endpoint: path,
		Observed: map[string]any{
			"react_lab_html":   path,
			"manual_evidence":  manualOut,
			"purpose":          "frontend API usage handoff",
			"renders":          []string{"localized functional areas", "evidence replay order", "Figma daemon boundary intent", "daily QA freshness", "manual QA workbench"},
			"source_evidence":  "ai-agent-product-acceptance",
			"i18n_source":      "contract.ui.i18n_dsl",
			"client_mutations": false,
		},
	}
}

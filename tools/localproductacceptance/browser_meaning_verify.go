package main

import (
	"os"
	"path/filepath"
	"strings"
)

func browserMeaningResult() browserMeaningProof {
	required := []string{
		"function matchesQuery",
		"JSON.stringify(scenario.observed || {})",
		"placeholder: t(\"toolbar.search\")",
		"developer_query_surface",
		"search_aliases",
		"search_entries",
		"aliases",
		"meaning",
	}
	return browserMeaningProof{Required: required, Missing: missingBrowserMeaningTokens(required)}
}

func missingBrowserMeaningTokens(required []string) []string {
	blob := browserMeaningSourceBlob()
	var missing []string
	for _, token := range required {
		if !strings.Contains(blob, token) {
			missing = append(missing, token)
		}
	}
	return missing
}

func browserMeaningSourceBlob() string {
	root, err := repoRoot()
	if err != nil {
		root = "."
	}
	paths := []string{"tools/localproductacceptance/contract_lab_template.go", "tools/localproductacceptance/qa_system.go"}
	var b strings.Builder
	for _, path := range paths {
		body, err := os.ReadFile(filepath.Join(root, path))
		if err == nil {
			b.Write(body)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

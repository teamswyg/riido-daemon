package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func scanRedactionDocs(root string) ([]string, error) {
	var problems []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil || entry.IsDir() || !isMarkdown(rel) {
			return err
		}
		if !isSecurityHubDoc(rel) && !isSecurityRedactionDoc(rel) {
			return nil
		}
		text, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		problems = append(problems, validateDoc(rel, string(text))...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan redaction docs: %w", err)
	}
	return problems, nil
}

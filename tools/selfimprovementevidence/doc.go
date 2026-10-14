package main

import (
	"fmt"
	"os"
)

func writeDoc(path string, m manifest) error {
	return os.WriteFile(path, []byte(renderMarkdown(m)), 0o644)
}

func checkDoc(path string, m manifest) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(body) != renderMarkdown(m) {
		return fmt.Errorf("generated doc drift: run go run ./tools/selfimprovementevidence -write-doc")
	}
	return nil
}

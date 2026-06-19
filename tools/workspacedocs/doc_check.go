package main

import (
	"fmt"
	"path/filepath"
)

func checkDocs(repo string, m manifest) []string {
	expected := renderRoot(m)
	path := filepath.Join(repo, m.GeneratedDoc)
	actual, err := readString(path)
	if err != nil {
		return []string{err.Error()}
	}
	if actual != expected {
		return []string{"generated doc drift: run go run ./tools/workspacedocs -write-doc"}
	}
	return nil
}

func writeDocs(repo string, m manifest) error {
	path := filepath.Join(repo, m.GeneratedDoc)
	if err := writeFile(path, renderRoot(m)); err != nil {
		return fmt.Errorf("write generated doc: %w", err)
	}
	return nil
}

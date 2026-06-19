package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func run(repoRoot, manifestPath, evidenceOut string, writeDoc, checkDoc, runIntegration bool) error {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}
	loaded, err := loadManifest(resolvePath(root, manifestPath))
	if err != nil {
		return err
	}
	if problems := validateManifest(root, loaded); len(problems) > 0 {
		return fmt.Errorf("provider observation manifest invalid:\n%s", joinProblems(problems))
	}
	if err := maybeWriteOrCheckDoc(root, loaded, writeDoc, checkDoc); err != nil {
		return err
	}
	if evidenceOut == "" && !runIntegration {
		return nil
	}
	evidence, err := observeProviders(root, loaded, runIntegration)
	if evidenceOut != "" {
		if writeErr := writeJSON(resolvePath(root, evidenceOut), evidence); writeErr != nil {
			return writeErr
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func maybeWriteOrCheckDoc(root string, loaded manifest, writeDoc, checkDoc bool) error {
	rendered := renderMarkdown(loaded)
	docPath := resolvePath(root, loaded.GeneratedDoc)
	if writeDoc {
		return os.WriteFile(docPath, []byte(rendered), 0o644)
	}
	if !checkDoc {
		return nil
	}
	current, err := os.ReadFile(docPath)
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(current) != rendered {
		return fmt.Errorf("generated doc drift: run tools/providerintegrationevidence -write-doc")
	}
	return nil
}

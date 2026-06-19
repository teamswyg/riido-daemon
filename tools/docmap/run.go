package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func run(repoRoot, manifestPath, evidenceOut string, write, check bool) error {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}
	loaded, err := loadManifest(resolvePath(root, manifestPath))
	if err != nil {
		return err
	}
	if problems := validateManifest(root, loaded); len(problems) > 0 {
		return fmt.Errorf("doc map invalid:\n%s", joinProblems(problems))
	}
	if err := maybeWriteOrCheck(root, loaded, write, check); err != nil {
		return err
	}
	if evidenceOut != "" {
		return writeJSON(resolvePath(root, evidenceOut), buildEvidence(loaded))
	}
	return nil
}

func maybeWriteOrCheck(root string, m manifest, write, check bool) error {
	rendered := map[string]string{
		m.GeneratedDocs.Readme:      renderReadme(m),
		m.GeneratedDocs.DocumentMap: renderDocumentMap(m),
	}
	for path, text := range rendered {
		if write {
			if err := writeText(resolvePath(root, path), text); err != nil {
				return err
			}
		}
		if check {
			current, err := os.ReadFile(resolvePath(root, path))
			if err != nil {
				return fmt.Errorf("read generated doc: %w", err)
			}
			if string(current) != text {
				return fmt.Errorf("generated doc drift: run tools/docmap -write")
			}
		}
	}
	return nil
}

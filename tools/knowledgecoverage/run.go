package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func run(repoRoot, manifestPath, evidenceOut string, writeDoc, checkDoc bool) error {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}
	m, err := loadManifest(resolvePath(root, manifestPath))
	if err != nil {
		return err
	}
	if problems := validateManifest(root, m); len(problems) > 0 {
		return fmt.Errorf("knowledge coverage manifest invalid:\n- %s", joinProblems(problems))
	}
	docs, problems := scanDocs(root, m)
	if err := maybeWriteOrCheckDoc(root, m, docs, problems, writeDoc, checkDoc); err != nil {
		return err
	}
	if evidenceOut != "" {
		if err := writeJSON(resolvePath(root, evidenceOut), buildEvidence(root, m, docs, problems)); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("knowledge coverage invalid:\n- %s", joinProblems(problems))
	}
	return nil
}

func maybeWriteOrCheckDoc(root string, m manifest, docs []docClass, problems []string, writeDoc, checkDoc bool) error {
	rendered := renderDoc(root, m, docs, problems)
	path := resolvePath(root, m.GeneratedDoc)
	if writeDoc {
		if err := writeText(path, rendered); err != nil {
			return err
		}
	}
	if checkDoc {
		current, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read generated doc: %w", err)
		}
		if string(current) != rendered {
			return fmt.Errorf("generated doc drift: run tools/knowledgecoverage -write-doc")
		}
	}
	return nil
}

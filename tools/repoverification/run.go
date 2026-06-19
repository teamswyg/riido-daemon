package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func run(repoRoot, manifestPath, evidenceOut string, writeDoc, checkDoc, shouldRun bool) error {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}
	loaded, err := loadManifest(resolvePath(root, manifestPath))
	if err != nil {
		return err
	}
	if problems := validateManifest(root, loaded); len(problems) > 0 {
		return fmt.Errorf("repo verification invalid:\n%s", joinProblems(problems))
	}
	if err := maybeWriteOrCheckDoc(root, loaded, writeDoc, checkDoc); err != nil {
		return err
	}
	var commandEvidence []commandEvidence
	if shouldRun {
		commandEvidence = runCommands(root, loaded.Commands)
	}
	if evidenceOut != "" {
		if err := writeJSON(resolvePath(root, evidenceOut), buildEvidence(loaded, commandEvidence)); err != nil {
			return err
		}
	}
	if anyFailed(commandEvidence) {
		return fmt.Errorf("one or more verification commands failed")
	}
	return nil
}

func maybeWriteOrCheckDoc(root string, loaded manifest, writeDoc, checkDoc bool) error {
	rendered := renderMarkdown(loaded)
	docPath := resolvePath(root, loaded.GeneratedDoc)
	if writeDoc {
		return writeText(docPath, rendered)
	}
	if !checkDoc {
		return nil
	}
	current, err := os.ReadFile(docPath)
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(current) != rendered {
		return fmt.Errorf("generated doc drift: run tools/repoverification -write-doc")
	}
	return nil
}

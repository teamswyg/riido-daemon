package main

import (
	"fmt"
	"path/filepath"
)

func run(repoRoot string) error {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}
	problems, err := scanRedactionDocs(root)
	if err != nil {
		return err
	}
	if len(problems) > 0 {
		return fmt.Errorf("redaction drift:\n%s", joinProblems(problems))
	}
	return nil
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func run(repoRoot, manifestPath, docPath string, write, check bool) error {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}
	loaded, err := loadManifest(resolvePath(root, manifestPath))
	if err != nil {
		return err
	}
	if docPath == "" {
		docPath = loaded.GeneratedDoc
	}
	if problems := validate(root, loaded); len(problems) > 0 {
		return fmt.Errorf("loop evidence invalid:\n%s", joinProblems(problems))
	}
	rendered := renderMarkdown(loaded)
	if write {
		return writeText(resolvePath(root, docPath), rendered)
	}
	if check {
		current, err := os.ReadFile(resolvePath(root, docPath))
		if err != nil {
			return fmt.Errorf("read generated doc: %w", err)
		}
		if string(current) != rendered {
			return fmt.Errorf("generated doc drift: run tools/loopevidence -write")
		}
	}
	return nil
}

func joinProblems(problems []string) string {
	var out strings.Builder
	for _, problem := range problems {
		out.WriteString("- ")
		out.WriteString(problem)
		out.WriteByte('\n')
	}
	return out.String()
}

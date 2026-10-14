package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func run(opts options) error {
	root, err := filepath.Abs(opts.Repo)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}
	loaded, err := loadManifest(resolvePath(root, opts.Manifest))
	if err != nil {
		return err
	}
	loaded, err = expandLoopFiles(root, loaded)
	if err != nil {
		return err
	}
	if opts.Doc == "" {
		opts.Doc = loaded.GeneratedDoc
	}
	problems := validate(root, loaded)
	rendered := ""
	if len(problems) == 0 {
		rendered = renderMarkdown(loaded)
	}
	if len(problems) == 0 && opts.Write {
		if err := writeText(resolvePath(root, opts.Doc), rendered); err != nil {
			problems = append(problems, err.Error())
		}
	}
	if len(problems) == 0 && !opts.Write && opts.Check {
		problems = append(problems, checkDoc(root, opts.Doc, rendered)...)
	}
	if opts.EvidenceOut != "" {
		if err := writeJSON(opts.EvidenceOut, buildEvidence(loaded, opts.Doc, problems)); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("loop evidence invalid:\n%s", joinProblems(problems))
	}
	return nil
}

func checkDoc(root, path, rendered string) []string {
	current, err := os.ReadFile(resolvePath(root, path))
	if err != nil {
		return []string{fmt.Sprintf("read generated doc: %v", err)}
	}
	if string(current) != rendered {
		return []string{"generated doc drift: run tools/loopevidence -write"}
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

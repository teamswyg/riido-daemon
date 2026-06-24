package main

import (
	"fmt"
	"os"
)

func run(opts options) error {
	root := repoRoot()
	reg, err := loadRegistry(repoPath(root, opts.Manifest))
	if err != nil {
		return err
	}
	problems := validateRegistry(root, reg)
	var changed *changedSummary
	if opts.ChangedFiles != "" {
		summary := changedCheck(root, reg, opts.ChangedFiles)
		changed = &summary
		problems = append(problems, summary.Problems...)
		if opts.GitHubAnnotations {
			emitGitHubAnnotations(summary)
		}
	}
	doc := renderDoc(reg)
	if opts.WriteDoc {
		if err := writeText(repoPath(root, reg.GeneratedDoc), doc); err != nil {
			return err
		}
	}
	if opts.CheckDoc {
		current, err := os.ReadFile(repoPath(root, reg.GeneratedDoc))
		if err != nil {
			return fmt.Errorf("read generated doc: %w", err)
		}
		if string(current) != doc {
			return fmt.Errorf("generated doc drift: run go run ./tools/loopregistry -write-doc")
		}
	}
	report := buildReport(reg, changed, problems)
	if opts.EvidenceOut != "" {
		if err := writeJSON(repoPath(root, opts.EvidenceOut), report); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("loop registry invalid:\n- %s", joinProblems(problems))
	}
	return nil
}

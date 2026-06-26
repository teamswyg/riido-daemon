package main

import (
	"fmt"
	"os"
)

func run(opts options) error {
	root := repoRoot()
	m, err := loadManifest(repoPath(root, opts.Manifest))
	if err != nil {
		return err
	}
	doc := renderDoc(m)
	if opts.WriteDoc {
		if err := writeText(repoPath(root, m.GeneratedDoc), doc); err != nil {
			return err
		}
	}
	if opts.CheckDoc {
		current, err := os.ReadFile(repoPath(root, m.GeneratedDoc))
		if err != nil {
			return fmt.Errorf("read generated doc: %w", err)
		}
		if string(current) != doc {
			return fmt.Errorf("generated doc drift: run go run ./tools/productloopevidence -write-doc")
		}
	}
	report, err := buildReport(root, m)
	if err != nil {
		return err
	}
	if opts.EvidenceOut != "" {
		if err := writeJSON(repoPath(root, opts.EvidenceOut), report); err != nil {
			return err
		}
	}
	if report.Status == statusFailed || opts.Strict && report.Status != statusPassed {
		return fmt.Errorf("product loop evidence status=%s problems=%v", report.Status, report.Problems)
	}
	return nil
}

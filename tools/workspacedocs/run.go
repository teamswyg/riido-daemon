package main

import (
	"fmt"
	"path/filepath"
)

func run(opts options) error {
	repo, err := filepath.Abs(opts.Repo)
	if err != nil {
		return err
	}
	m, err := readManifest(filepath.Join(repo, opts.Manifest))
	if err != nil {
		return err
	}
	problems, results := validateManifest(repo, m)
	if opts.WriteDoc && len(problems) == 0 {
		if err := writeDocs(repo, m); err != nil {
			return err
		}
	}
	if opts.CheckDoc {
		problems = append(problems, checkDocs(repo, m)...)
	}
	if err := writeJSON(opts.EvidenceOut, buildEvidence(m, results, problems)); err != nil {
		return err
	}
	if len(problems) > 0 {
		return fmt.Errorf("%s", problemText(problems))
	}
	return nil
}

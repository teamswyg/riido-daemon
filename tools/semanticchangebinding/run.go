package main

import (
	"context"
	"fmt"
)

func run(_ context.Context, opts options) error {
	manifest, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	problems := validateManifest(opts.Repo, manifest)
	changed := opts.ChangedFiles
	if changed == nil {
		changed = gitChangedFiles(opts.Repo)
	}
	results, bindingProblems := evaluate(manifest, changed)
	problems = append(problems, bindingProblems...)
	evidence := buildEvidence(manifest, changed, results, problems)
	if err := writeEvidence(opts.EvidenceOut, evidence); err != nil {
		return err
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("semantic-change-binding: clean")
	return nil
}

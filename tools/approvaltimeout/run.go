package main

import (
	"context"
	"fmt"
	"os"
)

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

func run(_ context.Context, opts options) error {
	manifest, err := loadJSON[Manifest](opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	problems, manifestChecks, sourceChecks := validate(opts.Repo, manifest)
	rendered := render(manifest)
	if err := maybeWriteDoc(opts, manifest, rendered); err != nil {
		return err
	}
	problems = append(problems, checkDoc(opts, manifest, rendered)...)
	if opts.EvidenceOut != "" {
		evidence := buildEvidence(manifest, problems, manifestChecks, sourceChecks)
		if err := writeJSON(opts.EvidenceOut, evidence); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("approval-wait-timeout: clean")
	return nil
}

func maybeWriteDoc(opts options, manifest Manifest, rendered string) error {
	if !opts.WriteDoc {
		return nil
	}
	path, err := cleanRepoPath(opts.Repo, manifest.GeneratedDoc)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(rendered), 0o644)
}

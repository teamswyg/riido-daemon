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
	manifest, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	rendered := renderedDocs(manifest)
	problems, sources, absent := validate(opts.Repo, manifest)
	if err := maybeWriteDocs(opts, rendered); err != nil {
		return err
	}
	problems = append(problems, checkDocs(opts, rendered)...)
	if opts.EvidenceOut != "" {
		evidence := buildEvidence(manifest, problems, sources, absent)
		if err := writeJSON(opts.EvidenceOut, evidence); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("saas-assignment-source: clean")
	return nil
}

func maybeWriteDocs(opts options, docs map[string]string) error {
	if !opts.WriteDoc {
		return nil
	}
	for rel, body := range docs {
		path, err := cleanRepoPath(opts.Repo, rel)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

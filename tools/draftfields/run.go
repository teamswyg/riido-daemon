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
	problems, allowed, forbidden := validate(opts.Repo, manifest)
	docs := renderAll(manifest)
	if err := maybeWriteDocs(opts, manifest, docs); err != nil {
		return err
	}
	problems = append(problems, checkDocs(opts, manifest, docs)...)
	if opts.EvidenceOut != "" {
		evidence := buildEvidence(manifest, problems, allowed, forbidden)
		if err := writeJSON(opts.EvidenceOut, evidence); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("draft-fields: clean")
	return nil
}

func maybeWriteDocs(opts options, manifest Manifest, docs renderedDocs) error {
	if !opts.WriteDoc {
		return nil
	}
	for _, doc := range docPairs(manifest, docs) {
		path, err := cleanRepoPath(opts.Repo, doc.path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(doc.body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

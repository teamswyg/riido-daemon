package main

import "fmt"

func run(opts options) error {
	opts = normalizeOptions(opts)
	m, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	docs := renderedDocs(m)
	problems := validateManifest(m)
	if err := maybeWriteDocs(opts, docs); err != nil {
		problems = append(problems, problem{Message: err.Error()})
	}
	problems = append(problems, checkDocs(opts, docs)...)
	if opts.EvidenceOut != "" {
		if err := writeJSON(opts.EvidenceOut, buildEvidence(m, problems)); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("figma-boundary-docs: clean")
	return nil
}

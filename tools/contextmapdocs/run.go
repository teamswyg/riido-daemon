package main

func run(opts options) error {
	m, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	problems := validateManifest(m)
	docs := renderedIfValid(m, problems)
	if err := maybeWriteDocs(opts, docs); err != nil {
		problems = append(problems, err.Error())
	}
	problems = append(problems, checkDocs(opts, docs)...)
	if opts.EvidenceOut != "" {
		if err := writeJSON(opts.EvidenceOut, buildEvidence(m, docs, problems)); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	return nil
}

func renderedIfValid(m manifest, problems []string) map[string]string {
	if len(problems) > 0 {
		return map[string]string{}
	}
	return renderedDocs(m)
}

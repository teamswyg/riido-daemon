package main

func run(opts options) error {
	m, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	problems, checks := validateManifest(opts.Repo, m)
	body := renderIfValid(m, problems)
	if err := maybeWriteDoc(opts, m, body); err != nil {
		problems = append(problems, err.Error())
	}
	problems = append(problems, checkDoc(opts, m, body)...)
	if opts.EvidenceOut != "" {
		if err := writeJSON(opts.EvidenceOut, buildEvidence(m, checks, problems)); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	return nil
}

func renderIfValid(m manifest, problems []string) string {
	if len(problems) > 0 {
		return ""
	}
	return renderDoc(m)
}

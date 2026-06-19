package main

import "fmt"

func run(opts options) error {
	opts = normalizeOptions(opts)
	m, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	sourceChecks := checkSources(opts.Repo, m.SourceChecks)
	problems := validateManifest(m)
	problems = append(problems, failedChecks("source check failed", sourceChecks)...)
	docs := map[string]string{}
	if len(problems) == 0 {
		docs = renderedDocs(m)
	}
	if err := maybeWriteDocs(opts, docs); err != nil {
		problems = append(problems, problem{Message: err.Error()})
	}
	problems = append(problems, checkDocs(opts, docs)...)
	if opts.EvidenceOut != "" {
		if err := writeJSON(opts.EvidenceOut, buildEvidence(m, problems, sourceChecks)); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("release-artifact-docs: clean")
	return nil
}

package main

import "fmt"

func run(opts options) error {
	m, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	c, err := loadContract(opts.Repo, m.StoreContract)
	if err != nil {
		return err
	}
	problems := validateInputs(m, c)
	docs := renderedIfValid(m, c, problems)
	if err := maybeWriteDocs(opts, docs); err != nil {
		problems = append(problems, err.Error())
	}
	problems = append(problems, checkDocs(opts, docs)...)
	if opts.EvidenceOut != "" {
		if err := writeJSON(opts.EvidenceOut, buildEvidence(m, c, docs, problems)); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("store-distribution-docs: clean")
	return nil
}

func renderedIfValid(m manifest, c contract, problems []string) map[string]string {
	if len(problems) > 0 {
		return map[string]string{}
	}
	return renderedDocs(m, c)
}

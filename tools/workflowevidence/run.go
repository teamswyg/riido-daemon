package main

import "fmt"

func run(opt options) error {
	root, err := findRepoRoot(opt.Repo)
	if err != nil {
		return err
	}
	m, err := loadManifest(repoPath(root, opt.Manifest))
	if err != nil {
		return err
	}
	result, err := auditWorkflows(root, m)
	if err != nil {
		return err
	}
	doc := renderDoc(m, result)
	if opt.WriteDoc {
		if err := writeText(repoPath(root, m.GeneratedDoc), doc); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if opt.CheckDoc {
		if err := verifyDoc(root, m, doc); err != nil {
			return err
		}
	}
	if opt.EvidenceOut != "" {
		if err := writeJSON(repoPath(root, opt.EvidenceOut), newEvidence(m, result)); err != nil {
			return err
		}
	}
	return verifyResult(result)
}

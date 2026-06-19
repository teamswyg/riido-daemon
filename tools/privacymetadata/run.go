package main

import (
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

func run(opts options) error {
	manifest, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	policy, err := loadPolicy(opts.Repo, manifest.PolicyArtifact)
	if err != nil {
		return err
	}
	problems, sources, shapes := validate(opts.Repo, manifest, policy)
	rendered := renderMarkdown(manifest, policy)
	if opts.WriteDoc {
		if err := writeText(repoPath(opts.Repo, manifest.GeneratedDoc), rendered); err != nil {
			return err
		}
	}
	problems = append(problems, checkDoc(opts, manifest.GeneratedDoc, rendered)...)
	if opts.EvidenceOut != "" {
		if err := writeJSON(opts.EvidenceOut, buildEvidence(manifest, policy, problems, sources, shapes)); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("privacy-metadata-evidence: clean")
	return nil
}

func checkDoc(opts options, rel, body string) []problem {
	if !opts.CheckDoc {
		return nil
	}
	current, err := os.ReadFile(repoPath(opts.Repo, rel))
	if err != nil {
		return []problem{{Message: err.Error()}}
	}
	if string(current) != body {
		return []problem{{Message: "generated doc drift: run tools/privacymetadata -write-doc"}}
	}
	return nil
}

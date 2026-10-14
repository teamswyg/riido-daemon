package main

import (
	"fmt"
	"os"
)

func run(opts options) error {
	opts = normalizeOptions(opts)
	manifest, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	problems, sources, behaviors := validate(opts.Repo, manifest)
	rendered := renderMarkdown(manifest)
	if opts.WriteDoc {
		if err := writeText(repoPath(opts.Repo, manifest.GeneratedDoc), rendered); err != nil {
			return err
		}
	}
	problems = append(problems, checkDoc(opts, manifest.GeneratedDoc, rendered)...)
	if opts.EvidenceOut != "" {
		ev := buildEvidence(manifest, problems, sources, behaviors)
		if err := writeJSON(opts.EvidenceOut, ev); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("executable-search-path-evidence: clean")
	return nil
}

func normalizeOptions(opts options) options {
	if opts.Repo == "" {
		opts.Repo = "."
	}
	if opts.Manifest == "" {
		opts.Manifest = defaultManifest
	}
	return opts
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
		return []problem{{Message: "generated doc drift: run tools/executablesearchpath -write-doc"}}
	}
	return nil
}

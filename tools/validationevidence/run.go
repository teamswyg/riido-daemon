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
	problems, sources, absent := validate(opts.Repo, manifest)
	rendered := renderMarkdown(manifest)
	if err := maybeWriteDoc(opts, manifest.GeneratedDoc, rendered); err != nil {
		return err
	}
	problems = append(problems, checkDoc(opts, manifest.GeneratedDoc, rendered)...)
	if opts.EvidenceOut != "" {
		if err := writeJSON(opts.EvidenceOut, buildEvidence(manifest, problems, sources, absent)); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("validation-evidence: clean")
	return nil
}

func maybeWriteDoc(opts options, rel, body string) error {
	if !opts.WriteDoc {
		return nil
	}
	return writeText(repoPath(opts.Repo, rel), body)
}

func checkDoc(opts options, rel, body string) []problem {
	if !opts.CheckDoc {
		return nil
	}
	current, err := os.ReadFile(repoPath(opts.Repo, rel))
	if err != nil {
		return []problem{{err.Error()}}
	}
	if string(current) != body {
		return []problem{{"generated doc drift: run tools/validationevidence -write-doc"}}
	}
	return nil
}

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
	if opts.WriteDoc {
		if err := writeText(repoPath(opts.Repo, manifest.GeneratedDoc), rendered); err != nil {
			return err
		}
	}
	problems = append(problems, checkDoc(opts, manifest.GeneratedDoc, rendered)...)
	if opts.EvidenceOut != "" {
		evidence := buildEvidence(manifest, problems, sources, absent)
		if err := writeJSON(opts.EvidenceOut, evidence); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("local-daemon-contract-evidence: clean")
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
		return []problem{{Message: "generated doc drift: run tools/localdaemoncontract -write-doc"}}
	}
	return nil
}

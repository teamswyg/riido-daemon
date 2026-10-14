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
	fsm := buildFSMSnapshot()
	rendered := renderMarkdown(manifest, fsm)
	problems, sources, forbiddenOK := validate(opts.Repo, manifest, rendered)
	if opts.WriteDoc {
		if err := writeText(repoPath(opts.Repo, manifest.GeneratedDoc), rendered); err != nil {
			return err
		}
	}
	problems = append(problems, checkDoc(opts, manifest.GeneratedDoc, rendered)...)
	if opts.EvidenceOut != "" {
		if err := writeJSON(opts.EvidenceOut, buildEvidence(manifest, fsm, problems, sources, forbiddenOK)); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("assignment-fsm-evidence: clean")
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
		return []problem{{Message: "generated doc drift: run tools/assignmentfsm -write-doc"}}
	}
	return nil
}

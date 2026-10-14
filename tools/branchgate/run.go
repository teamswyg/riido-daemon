package main

import (
	"context"
	"fmt"
)

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	WriteScript bool
	CheckScript bool
	EvidenceOut string
}

func run(_ context.Context, opts options) error {
	manifest, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	rendered := renderedFiles{Doc: renderDoc(manifest), Script: renderScript(manifest)}
	var problems []problem
	problems = append(problems, maybeWrite(opts, manifest, rendered)...)
	scriptChecks := checkGenerated(opts, manifest, rendered)
	problems = append(problems, scriptCheckProblems(scriptChecks)...)
	exampleProblems, examples := runExamples(opts.Repo, manifest)
	problems = append(problems, exampleProblems...)
	if opts.EvidenceOut != "" {
		evidence := buildEvidence(manifest, problems, scriptChecks, examples)
		if err := writeJSON(opts.EvidenceOut, evidence); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("branch-gate: clean")
	return nil
}

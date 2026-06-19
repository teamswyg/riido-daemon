package main

import "fmt"

func run(opts options) error {
	opts = normalizeOptions(opts)
	manifest, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	envConsts := checkEnvConstants(opts.Repo, manifest)
	anchors := checkAnchors(opts.Repo, manifest)
	sources := checkSources(opts.Repo, manifest)
	rendered := renderMarkdown(manifest)
	var problems []problem
	if err := maybeWriteDoc(opts, manifest, rendered); err != nil {
		problems = append(problems, problem{Message: err.Error()})
	}
	problems = append(problems, checkDoc(opts, manifest, rendered)...)
	problems = append(problems, resultProblems("env const check failed", envConsts)...)
	problems = append(problems, resultProblems("anchor check failed", anchors)...)
	problems = append(problems, resultProblems("source check failed", sources)...)
	if opts.EvidenceOut != "" {
		if err := writeJSON(opts.EvidenceOut, buildEvidence(manifest, problems, envConsts, anchors, sources)); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("config-reference-evidence: clean")
	return nil
}

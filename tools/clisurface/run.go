package main

import "fmt"

func run(opts options) error {
	opts = normalizeOptions(opts)
	manifest, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	sources := checkSources(opts.Repo, manifest)
	forbidden := checkForbidden(opts.Repo, manifest)
	behaviors := checkBehaviors(opts.Repo, manifest)
	rendered := renderMarkdown(manifest)
	var problems []problem
	if err := maybeWriteDoc(opts, manifest, rendered); err != nil {
		problems = append(problems, problem{Message: err.Error()})
	}
	problems = append(problems, checkDoc(opts, manifest, rendered)...)
	problems = appendFailedProblems(problems, "source check failed", sources)
	problems = appendFailedProblems(problems, "forbidden source token", forbidden)
	problems = appendFailedProblems(problems, "behavior check failed", behaviors)
	if opts.EvidenceOut != "" {
		ev := buildEvidence(manifest, problems, sources, forbidden, behaviors)
		if err := writeJSON(opts.EvidenceOut, ev); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("cli-surface-evidence: clean")
	return nil
}

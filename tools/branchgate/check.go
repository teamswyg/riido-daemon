package main

import "strings"

func checkGenerated(opts options, manifest Manifest, rendered renderedFiles) []ScriptCheck {
	checks := []ScriptCheck{}
	if opts.CheckDoc {
		checks = append(checks, checkOne(opts.Repo, "reader doc", manifest.GeneratedDoc, rendered.Doc))
	}
	if opts.CheckScript {
		checks = append(checks, checkOne(opts.Repo, "branch script", manifest.GeneratedScript, rendered.Script))
	}
	checks = append(checks, checkContains(opts.Repo, "PR workflow calls script", manifest.Workflow, manifest.GeneratedScript))
	checks = append(checks, checkContains(opts.Repo, "evidence workflow runs tool", manifest.EvidenceWorkflow, "go run ./tools/branchgate"))
	return checks
}

func checkOne(repo, name, rel, want string) ScriptCheck {
	actual, err := readFile(repo, rel)
	return ScriptCheck{Name: name, File: rel, Pass: err == nil && actual == want}
}

func checkContains(repo, name, rel, want string) ScriptCheck {
	if rel == "" {
		return ScriptCheck{Name: name, Pass: true}
	}
	actual, err := readFile(repo, rel)
	return ScriptCheck{Name: name, File: rel, Pass: err == nil && strings.Contains(actual, want)}
}

func scriptCheckProblems(checks []ScriptCheck) []problem {
	var problems []problem
	for _, check := range checks {
		if !check.Pass {
			problems = append(problems, problem{Message: "generated drift: " + check.File})
		}
	}
	return problems
}

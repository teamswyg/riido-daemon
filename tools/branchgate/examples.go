package main

import (
	"errors"
	"os/exec"
)

func runExamples(repo string, manifest Manifest) ([]problem, []ExampleCheck) {
	var problems []problem
	checks := make([]ExampleCheck, 0, len(manifest.Examples))
	script, err := cleanRepoPath(repo, manifest.GeneratedScript)
	if err != nil {
		return []problem{{err.Error()}}, checks
	}
	for _, example := range manifest.Examples {
		check := runExample(script, example)
		if !check.Pass {
			problems = append(problems, problem{Message: "branch example drift: " + example.Branch})
		}
		checks = append(checks, check)
	}
	return problems, checks
}

func runExample(script string, example BranchExample) ExampleCheck {
	cmd := exec.Command("bash", script, example.Branch)
	err := cmd.Run()
	code := exitCode(err)
	pass := (code == 0) == example.Accepted
	return ExampleCheck{Branch: example.Branch, WantAccepted: example.Accepted, ExitCode: code, Pass: pass}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

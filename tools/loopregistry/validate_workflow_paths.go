package main

import (
	"os"
)

func validateWorkflowPathCoverage(root string, reg registry) []string {
	body, err := os.ReadFile(repoPath(root, reg.Workflow))
	if err != nil {
		return []string{"workflow file missing: " + reg.Workflow}
	}
	patterns := workflowPathPatterns(string(body))
	var problems []string
	for _, claim := range reg.BusinessClaims {
		for _, rel := range boundClaimPaths(claim) {
			if !workflowCoversPath(patterns, rel) {
				problems = append(problems, claim.ID+" bound path is not covered by workflow paths: "+rel)
			}
		}
	}
	return problems
}

func boundClaimPaths(claim businessClaim) []string {
	out := append([]string{}, claim.Files...)
	out = append(out, claim.Docs...)
	for _, check := range append(claim.Verifiers, claim.Contracts...) {
		out = append(out, check.File)
	}
	return out
}

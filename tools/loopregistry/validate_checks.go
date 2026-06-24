package main

import (
	"os"
	"strings"
)

func validateChecks(root, claimID string, checks []sourceCheck, kind string) []string {
	if len(checks) == 0 && kind == "verifier" {
		return []string{claimID + " requires verifier checks"}
	}
	var problems []string
	for _, check := range checks {
		problems = append(problems, validateCheck(root, claimID, check, kind)...)
	}
	return problems
}

func validateCheck(root, claimID string, check sourceCheck, kind string) []string {
	if check.Name == "" || check.File == "" || len(check.Contains) == 0 {
		return []string{claimID + " has incomplete " + kind + " check"}
	}
	body, err := os.ReadFile(repoPath(root, check.File))
	if err != nil {
		return []string{claimID + " " + kind + " file missing: " + check.File}
	}
	text := string(body)
	var missing []string
	for _, needle := range check.Contains {
		if !strings.Contains(text, needle) {
			missing = append(missing, needle)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return []string{claimID + " " + kind + " " + check.Name + " missing tokens: " + strings.Join(missing, ", ")}
}

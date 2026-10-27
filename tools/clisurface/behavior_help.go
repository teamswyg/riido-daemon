package main

import "strings"

func checkHelpOutput(repo string, manifest Manifest) CheckResult {
	out, err := runGoCommand(repo, "run", "./cmd/riido", "--help")
	result := CheckResult{Name: "help-output", Command: "go run ./cmd/riido --help", Pass: err == nil}
	if err != nil {
		result.Detail = err.Error()
		return result
	}
	result.Detail = checkRootHelpTokens(out, manifest.CommandGroups)
	result.Pass = result.Detail == ""
	return result
}

func checkRootHelpTokens(out string, groups []CommandGroup) string {
	if strings.Contains(out, "riido claude-permission-mcp") {
		return "internal permission MCP command exposed in root help"
	}
	for _, group := range groups {
		if group.RootHelp != "public" && group.RootHelp != "internal" {
			return "invalid root-help visibility: " + group.Name
		}
		if group.RootHelp == "internal" && strings.Contains(out, "riido "+group.Name) {
			return "internal command exposed: " + group.Name
		}
		for _, token := range group.UsageTokens {
			present := strings.Contains(out, token)
			if group.RootHelp == "public" && !present {
				return "missing public usage token: " + token
			}
		}
	}
	return ""
}

func checkInternalHelpOutput(repo string, manifest Manifest) CheckResult {
	out, err := runGoCommand(repo, "run", "./cmd/riido", "daemon", "--help")
	result := CheckResult{Name: "internal-help-output", Command: "go run ./cmd/riido daemon --help", Pass: err == nil}
	if err != nil {
		result.Detail = err.Error()
		return result
	}
	for _, group := range manifest.CommandGroups {
		for _, token := range group.UsageTokens {
			if !strings.Contains(out, token) {
				result.Pass, result.Detail = false, "missing legacy usage token: "+token
				return result
			}
		}
	}
	return result
}

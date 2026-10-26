package claude

import "strings"

func appendContextCLIReadArgs(args []string, enabled bool) []string {
	if !enabled {
		return args
	}
	var rules []string
	for _, leaf := range []string{"context get", "work get", "work tree", "work relations", "work comments", "agents list", "agents show"} {
		rules = append(rules, "Bash(riido "+leaf+")", "Bash(riido "+leaf+" *)")
	}
	return append(args, "--allowedTools", strings.Join(rules, ","))
}

func startBlockedArgs(contextCLIReads bool) []string {
	blocked := BlockedArgs()
	if contextCLIReads {
		blocked = append(blocked, "--allowedTools", "--allowed-tools")
	}
	return blocked
}

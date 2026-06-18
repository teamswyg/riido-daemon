package cursor

import "strings"

func pickProfile(help string) Profile {
	lower := strings.ToLower(help)
	switch {
	case hasSubcommand(lower, "chat"):
		return ProfileLegacyChat
	case hasSubcommand(lower, "agent"):
		return ProfileAgentSubcommand
	default:
		return ProfileRootPrint
	}
}

func hasSubcommand(help, name string) bool {
	return strings.Contains(help, "\n"+name) ||
		strings.Contains(help, "  "+name+" ") ||
		strings.Contains(help, " "+name+"  ")
}

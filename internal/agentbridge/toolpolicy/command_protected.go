package toolpolicy

import "strings"

func commandTouchesProtectedPath(command string) bool {
	normalized := strings.ToLower(command)
	if !commandMentionsProtectedPath(normalized) {
		return false
	}
	for _, marker := range []string{">", " rm ", "rm -", " mv ", " cp ", "tee ", "sed -i", "perl -pi", "chmod ", "chown "} {
		if strings.Contains(" "+normalized, marker) {
			return true
		}
	}
	return false
}

func commandMentionsProtectedPath(command string) bool {
	for _, token := range protectedCommandPathTokens() {
		if strings.Contains(command, token) {
			return true
		}
	}
	return false
}

func protectedCommandPathTokens() []string {
	return []string{
		".git", ".vscode", ".idea", ".husky", ".claude", ".env", ".aws",
		".ssh", ".gnupg", ".docker/config.json", ".config/gh/hosts.yml",
	}
}

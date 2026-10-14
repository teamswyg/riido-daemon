package toolpolicy

import "strings"

func commandMentionsSecretPath(command string) bool {
	for _, token := range secretPathTokens() {
		if strings.Contains(command, token) {
			return true
		}
	}
	return false
}

func secretPathTokens() []string {
	return []string{
		".env", ".aws/credentials", ".aws/config", ".ssh/", ".gnupg/",
		".npmrc", ".pypirc", ".netrc", ".docker/config.json",
		".config/gh/hosts.yml",
	}
}

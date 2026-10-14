package toolpolicy

import "strings"

func commandExposesSecrets(command string) bool {
	normalized := strings.ToLower(command)
	if commandContainsSecretManagerAccess(normalized) {
		return true
	}
	if !commandMentionsSecretPath(normalized) {
		return false
	}
	return commandContainsReadMarker(normalized)
}

func commandContainsSecretManagerAccess(command string) bool {
	for _, marker := range []string{
		"aws secretsmanager get-secret-value",
		"aws ssm get-parameter",
		"gh auth token",
		"security find-generic-password",
	} {
		if strings.Contains(command, marker) {
			return true
		}
	}
	return false
}

func commandContainsReadMarker(command string) bool {
	for _, marker := range []string{" cat ", " grep ", " rg ", " awk ", " sed ", " less ", " more ", " head ", " tail "} {
		if strings.Contains(" "+command, marker) {
			return true
		}
	}
	return false
}

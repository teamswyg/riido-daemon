package toolpolicy

import (
	"strings"
)

func commandArg(args map[string]string) (string, bool) {
	for _, key := range []string{"command", "cmd", "script", "input.command"} {
		if value, ok := args[key]; ok {
			return value, true
		}
	}
	for key, value := range args {
		normalized := normalizeToolToken(key)
		if normalized == "command" || normalized == "cmd" || strings.HasSuffix(normalized, "_command") {
			return value, true
		}
	}
	return "", false
}

func commandContainsNetworkEgress(command string) bool {
	normalized := strings.ToLower(command)
	return strings.Contains(normalized, "http://") ||
		strings.Contains(normalized, "https://") ||
		strings.Contains(normalized, "curl ") ||
		strings.Contains(normalized, "wget ") ||
		strings.Contains(normalized, "nc ") ||
		strings.Contains(normalized, "netcat ")
}

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

func commandExposesSecrets(command string) bool {
	normalized := strings.ToLower(command)
	for _, marker := range []string{
		"aws secretsmanager get-secret-value",
		"aws ssm get-parameter",
		"gh auth token",
		"security find-generic-password",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	if !commandMentionsSecretPath(normalized) {
		return false
	}
	for _, marker := range []string{" cat ", " grep ", " rg ", " awk ", " sed ", " less ", " more ", " head ", " tail "} {
		if strings.Contains(" "+normalized, marker) {
			return true
		}
	}
	return false
}

func commandIsDestructive(command string) bool {
	normalized := strings.ToLower(strings.TrimSpace(command))
	for _, marker := range []string{
		"rm -rf",
		"rm -fr",
		"sudo ",
		"chmod 777",
		"chown ",
		"dd if=",
		"dd of=",
		"mkfs",
		"git reset --hard",
		"git clean -fd",
		"git push",
		"terraform apply",
		"terraform destroy",
		"kubectl delete",
		"aws cloudformation delete",
		"aws dynamodb delete",
		"aws ecr delete",
		"aws iam delete",
		"aws s3 rm",
		"aws secretsmanager delete",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func commandMentionsProtectedPath(command string) bool {
	for _, token := range []string{".git", ".vscode", ".idea", ".husky", ".claude", ".env", ".aws", ".ssh", ".gnupg", ".docker/config.json", ".config/gh/hosts.yml"} {
		if strings.Contains(command, token) {
			return true
		}
	}
	return false
}

func commandMentionsSecretPath(command string) bool {
	for _, token := range []string{".env", ".aws/credentials", ".aws/config", ".ssh/", ".gnupg/", ".npmrc", ".pypirc", ".netrc", ".docker/config.json", ".config/gh/hosts.yml"} {
		if strings.Contains(command, token) {
			return true
		}
	}
	return false
}

func isProtectedPath(path string) bool {
	path = normalizePath(path)
	if path == "" {
		return false
	}
	if path == ".env" || strings.HasPrefix(path, ".env.") || strings.HasPrefix(path, ".env/") {
		return true
	}
	for _, file := range []string{
		".bash_profile",
		".bashrc",
		".claude.json",
		".gitconfig",
		".gitmodules",
		".mcp.json",
		".netrc",
		".npmrc",
		".profile",
		".pypirc",
		".ripgreprc",
		".zprofile",
		".zshrc",
		"credentials",
	} {
		if path == file || strings.HasSuffix(path, "/"+file) {
			return true
		}
	}
	if strings.HasPrefix(path, ".claude/commands/") ||
		strings.HasPrefix(path, ".claude/agents/") ||
		strings.HasPrefix(path, ".claude/skills/") ||
		strings.HasPrefix(path, ".claude/worktrees/") {
		return false
	}
	for _, dir := range []string{".git", ".vscode", ".idea", ".husky", ".claude", ".aws", ".ssh", ".gnupg"} {
		if path == dir || strings.HasPrefix(path, dir+"/") || strings.Contains(path, "/"+dir+"/") {
			return true
		}
	}
	if path == ".docker/config.json" || strings.HasSuffix(path, "/.docker/config.json") {
		return true
	}
	if path == ".config/gh/hosts.yml" || strings.HasSuffix(path, "/.config/gh/hosts.yml") {
		return true
	}
	return false
}

func normalizePath(path string) string {
	path = strings.ToLower(strings.TrimSpace(path))
	path = strings.Trim(path, `"'`)
	path = strings.ReplaceAll(path, "\\", "/")
	for strings.HasPrefix(path, "./") {
		path = strings.TrimPrefix(path, "./")
	}
	return path
}

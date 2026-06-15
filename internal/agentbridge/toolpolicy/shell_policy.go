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
	if !strings.Contains(normalized, ".git") && !strings.Contains(normalized, ".vscode") && !strings.Contains(normalized, ".idea") && !strings.Contains(normalized, ".husky") && !strings.Contains(normalized, ".claude") {
		return false
	}
	for _, marker := range []string{">", " rm ", "rm -", " mv ", " cp ", "tee ", "sed -i", "perl -pi", "chmod ", "chown "} {
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

func isProtectedPath(path string) bool {
	path = normalizePath(path)
	if path == "" {
		return false
	}
	for _, file := range []string{
		".bash_profile",
		".bashrc",
		".claude.json",
		".gitconfig",
		".gitmodules",
		".mcp.json",
		".profile",
		".ripgreprc",
		".zprofile",
		".zshrc",
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
	for _, dir := range []string{".git", ".vscode", ".idea", ".husky", ".claude"} {
		if path == dir || strings.HasPrefix(path, dir+"/") || strings.Contains(path, "/"+dir+"/") {
			return true
		}
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

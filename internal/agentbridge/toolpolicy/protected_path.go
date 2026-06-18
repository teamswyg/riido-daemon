package toolpolicy

import "strings"

func isProtectedPath(path string) bool {
	path = normalizePath(path)
	if path == "" {
		return false
	}
	if isProtectedEnvPath(path) || isProtectedNamedFile(path) {
		return true
	}
	if isAllowedClaudePath(path) {
		return false
	}
	return isProtectedDirectoryPath(path) || isProtectedConfigPath(path)
}

func isProtectedEnvPath(path string) bool {
	return path == ".env" || strings.HasPrefix(path, ".env.") || strings.HasPrefix(path, ".env/")
}

func isProtectedNamedFile(path string) bool {
	for _, file := range protectedFileNames() {
		if path == file || strings.HasSuffix(path, "/"+file) {
			return true
		}
	}
	return false
}

func isAllowedClaudePath(path string) bool {
	return strings.HasPrefix(path, ".claude/commands/") ||
		strings.HasPrefix(path, ".claude/agents/") ||
		strings.HasPrefix(path, ".claude/skills/") ||
		strings.HasPrefix(path, ".claude/worktrees/")
}

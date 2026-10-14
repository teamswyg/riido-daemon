package toolpolicy

import "strings"

func isProtectedDirectoryPath(path string) bool {
	for _, dir := range protectedDirectoryNames() {
		if path == dir || strings.HasPrefix(path, dir+"/") || strings.Contains(path, "/"+dir+"/") {
			return true
		}
	}
	return false
}

func isProtectedConfigPath(path string) bool {
	if path == ".docker/config.json" || strings.HasSuffix(path, "/.docker/config.json") {
		return true
	}
	return path == ".config/gh/hosts.yml" || strings.HasSuffix(path, "/.config/gh/hosts.yml")
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

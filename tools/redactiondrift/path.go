package main

import (
	"path/filepath"
	"strings"
)

func slash(path string) string {
	return filepath.ToSlash(path)
}

func isMarkdown(path string) bool {
	return strings.HasSuffix(path, ".md")
}

func isSecurityRedactionDoc(path string) bool {
	path = slash(path)
	return path == "docs/20-domain/security-redaction.md" ||
		strings.HasPrefix(path, "docs/20-domain/security-redaction/")
}

func isSecurityHubDoc(path string) bool {
	path = slash(path)
	return path == "docs/20-domain/security.md" ||
		strings.HasPrefix(path, "docs/20-domain/security/")
}

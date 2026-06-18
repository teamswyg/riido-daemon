package workdir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func safeJoin(root, rel string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", errors.New("workdir: root is required")
	}
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return "", errors.New("workdir: relative path is required")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("workdir: relative path is absolute: %s", rel)
	}
	clean := filepath.Clean(rel)
	if pathEscapesRoot(clean) {
		return "", fmt.Errorf("workdir: relative path escapes root: %s", rel)
	}
	return filepath.Join(root, clean), nil
}

func pathEscapesRoot(clean string) bool {
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return true
	}
	return slices.Contains(strings.Split(filepath.ToSlash(clean), "/"), "..")
}

package workdir

import (
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func sortedUniquePaths(paths []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		path = filepath.ToSlash(strings.TrimSpace(path))
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

func safePathSegment(s string) bool {
	if s == "" {
		return false
	}
	if strings.ContainsRune(s, os.PathSeparator) {
		return false
	}
	if strings.Contains(s, "..") {
		return false
	}
	return true
}

func localFileURI(path string) string {
	return (&url.URL{Scheme: "file", Path: path}).String()
}

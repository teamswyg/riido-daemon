package detectutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Mode().IsRegular()
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || !info.Mode().IsRegular() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode().Perm()&0o111 != 0
}

func executableNames(name string) []string {
	if runtime.GOOS != "windows" || filepath.Ext(name) != "" {
		return []string{name}
	}
	return windowsExecutableNames(name)
}

func windowsExecutableNames(name string) []string {
	exts := filepath.SplitList(os.Getenv("PATHEXT"))
	if len(exts) == 0 {
		exts = []string{".COM", ".EXE", ".BAT", ".CMD"}
	}
	out := make([]string, 0, len(exts))
	for _, ext := range exts {
		if ext = strings.TrimSpace(ext); ext != "" {
			out = append(out, name+ext)
		}
	}
	return fallbackExecutableNames(name, out)
}

func fallbackExecutableNames(name string, values []string) []string {
	if len(values) == 0 {
		return []string{name}
	}
	return values
}

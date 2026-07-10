package detectutil

import "path/filepath"

func macOSAppResourceDirs(home string) []string {
	dirs := []string{
		"/Applications/ChatGPT.app/Contents/Resources",
		"/Applications/Codex.app/Contents/Resources",
	}
	if home == "" {
		return dirs
	}
	return append(dirs,
		filepath.Join(home, "Applications", "ChatGPT.app", "Contents", "Resources"),
		filepath.Join(home, "Applications", "Codex.app", "Contents", "Resources"),
	)
}

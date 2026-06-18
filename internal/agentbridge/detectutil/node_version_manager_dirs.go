package detectutil

import "path/filepath"

func nodeVersionManagerBins(home string) []string {
	var out []string
	for _, glob := range nodeVersionManagerGlobs(home) {
		matches, err := filepath.Glob(glob)
		if err == nil {
			out = append(out, matches...)
		}
	}
	return out
}

func nodeVersionManagerGlobs(home string) []string {
	return []string{
		filepath.Join(home, ".nvm", "versions", "node", "*", "bin"),
		filepath.Join(home, ".fnm", "node-versions", "*", "installation", "bin"),
		filepath.Join(home, "Library", "Application Support", "fnm", "node-versions", "*", "installation", "bin"),
		filepath.Join(home, ".local", "share", "fnm", "node-versions", "*", "installation", "bin"),
		filepath.Join(home, ".asdf", "installs", "nodejs", "*", "bin"),
	}
}

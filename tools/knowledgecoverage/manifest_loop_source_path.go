package main

import "path/filepath"

func manifestSiblingSourcePath(root, ownerPath, source string) (string, bool) {
	if filepath.IsAbs(source) {
		return manifestSourcePath(root, source)
	}
	return manifestSourcePath(root, filepath.Join(filepath.Dir(ownerPath), filepath.FromSlash(source)))
}

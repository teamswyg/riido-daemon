package main

import (
	"fmt"
	"path/filepath"
	"sort"
)

func validateLoopFileInventory(root string, files []string) []string {
	if len(files) == 0 {
		return nil
	}
	listed := map[string]bool{}
	dirs := map[string]bool{}
	for _, file := range files {
		listed[file] = true
		dirs[filepath.ToSlash(filepath.Dir(file))] = true
	}
	var problems []string
	for dir := range dirs {
		matches, err := filepath.Glob(resolvePath(root, dir+"/*.riido.json"))
		if err != nil {
			problems = append(problems, fmt.Sprintf("scan loop file dir %q: %v", dir, err))
			continue
		}
		for _, match := range matches {
			rel, err := filepath.Rel(root, match)
			if err != nil {
				problems = append(problems, fmt.Sprintf("rel loop file %q: %v", match, err))
				continue
			}
			rel = filepath.ToSlash(rel)
			if !listed[rel] {
				problems = append(problems, "unregistered loop file "+rel)
			}
		}
	}
	sort.Strings(problems)
	return problems
}

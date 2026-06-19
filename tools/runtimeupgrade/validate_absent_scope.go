package main

import "os"

func scanAbsentScope(repo, scope string, tokens []string, check *AbsentCheck, problems *[]problem) {
	root, err := cleanRepoPath(repo, scope)
	if err != nil {
		*problems = append(*problems, problem{err.Error()})
		return
	}
	info, err := os.Stat(root)
	if err != nil {
		*problems = append(*problems, problem{err.Error()})
		return
	}
	if !info.IsDir() {
		scanAbsentFile(repo, root, tokens, check)
		return
	}
	walkAbsentDir(repo, root, tokens, check, problems)
}

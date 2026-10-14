package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func scanAbsentPath(repo, path string, tokens []string, check *AbsentCheck, problems *[]problem) {
	info, err := os.Stat(path)
	if err != nil {
		*problems = append(*problems, problem{err.Error()})
		return
	}
	if !info.IsDir() {
		scanAbsentFile(repo, path, tokens, check)
		return
	}
	err = filepath.WalkDir(path, func(item string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		scanAbsentFile(repo, item, tokens, check)
		return nil
	})
	if err != nil {
		*problems = append(*problems, problem{err.Error()})
	}
}

func scanAbsentFile(repo, path string, tokens []string, check *AbsentCheck) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, token := range tokens {
		if strings.Contains(string(data), token) {
			check.Hits = append(check.Hits, absentHit(repo, path, token))
		}
	}
}

func absentHit(repo, path, token string) string {
	rel, err := filepath.Rel(repo, path)
	if err != nil {
		rel = path
	}
	return fmt.Sprintf("%s:%s", rel, token)
}

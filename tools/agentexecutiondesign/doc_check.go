package main

import (
	"fmt"
	"os"
)

func maybeWriteDocs(opts options, docs map[string]string) error {
	if !opts.WriteDoc {
		return nil
	}
	for path, body := range docs {
		if err := writeText(repoPath(opts.Repo, path), body); err != nil {
			return err
		}
	}
	return nil
}

func checkDocs(opts options, docs map[string]string) []string {
	if !opts.CheckDoc {
		return nil
	}
	var problems []string
	for path, expected := range docs {
		data, err := os.ReadFile(repoPath(opts.Repo, path))
		if err != nil {
			problems = append(problems, fmt.Sprintf("read %s: %v", path, err))
			continue
		}
		if string(data) != expected {
			problems = append(problems, fmt.Sprintf("%s is stale; run agentexecutiondesign -write-doc", path))
		}
	}
	return problems
}

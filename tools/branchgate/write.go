package main

import (
	"os"
	"path/filepath"
)

func maybeWrite(opts options, manifest Manifest, rendered renderedFiles) []problem {
	var problems []problem
	if opts.WriteDoc {
		problems = append(problems, writeOne(opts.Repo, manifest.GeneratedDoc, rendered.Doc, 0o644)...)
	}
	if opts.WriteScript {
		problems = append(problems, writeOne(opts.Repo, manifest.GeneratedScript, rendered.Script, 0o755)...)
	}
	return problems
}

func writeOne(repo, rel, body string, perm os.FileMode) []problem {
	path, err := cleanRepoPath(repo, rel)
	if err != nil {
		return []problem{{err.Error()}}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return []problem{{err.Error()}}
	}
	if err := os.WriteFile(path, []byte(body), perm); err != nil {
		return []problem{{err.Error()}}
	}
	return nil
}

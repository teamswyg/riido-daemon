package main

import "os"

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

func checkDocs(opts options, docs map[string]string) []problem {
	if !opts.CheckDoc {
		return nil
	}
	var problems []problem
	for path, body := range docs {
		data, err := os.ReadFile(repoPath(opts.Repo, path))
		if err != nil {
			problems = append(problems, problem{Message: err.Error()})
			continue
		}
		if string(data) != body {
			problems = append(problems, problem{Message: "generated doc drift: run go run ./tools/integrationmatrix -write-doc"})
		}
	}
	return problems
}

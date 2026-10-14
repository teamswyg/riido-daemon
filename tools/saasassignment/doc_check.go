package main

import "os"

func checkDocs(opts options, docs map[string]string) []problem {
	if !opts.CheckDoc {
		return nil
	}
	var problems []problem
	for rel, rendered := range docs {
		path, err := cleanRepoPath(opts.Repo, rel)
		if err != nil {
			problems = append(problems, problem{err.Error()})
			continue
		}
		actual, err := os.ReadFile(path)
		if err != nil {
			problems = append(problems, problem{err.Error()})
			continue
		}
		if string(actual) != rendered {
			problems = append(problems, problem{"generated doc drift: run go run ./tools/saasassignment -write-doc"})
		}
	}
	return problems
}

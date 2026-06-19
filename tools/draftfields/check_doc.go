package main

import "os"

func checkDocs(opts options, manifest Manifest, docs renderedDocs) []problem {
	if !opts.CheckDoc {
		return nil
	}
	var problems []problem
	for _, doc := range docPairs(manifest, docs) {
		problems = append(problems, checkOneDoc(opts.Repo, doc)...)
	}
	return problems
}

func checkOneDoc(repo string, doc renderedDoc) []problem {
	path, err := cleanRepoPath(repo, doc.path)
	if err != nil {
		return []problem{{err.Error()}}
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		return []problem{{err.Error()}}
	}
	if string(actual) != doc.body {
		return []problem{{"generated doc drift: run go run ./tools/draftfields -write-doc"}}
	}
	return nil
}

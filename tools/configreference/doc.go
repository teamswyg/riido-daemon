package main

import "os"

func maybeWriteDoc(opts options, manifest Manifest, rendered string) error {
	if !opts.WriteDoc {
		return nil
	}
	for path, body := range renderedDocs(manifest, rendered) {
		if err := writeText(repoPath(opts.Repo, path), body); err != nil {
			return err
		}
	}
	return nil
}

func checkDoc(opts options, manifest Manifest, rendered string) []problem {
	if !opts.CheckDoc {
		return nil
	}
	var problems []problem
	for path, body := range renderedDocs(manifest, rendered) {
		data, err := os.ReadFile(repoPath(opts.Repo, path))
		if err != nil {
			problems = append(problems, problem{Message: err.Error()})
			continue
		}
		if string(data) != body {
			problems = append(problems, problem{Message: "generated doc drift: run go run ./tools/configreference -write-doc"})
		}
	}
	return problems
}

func renderedDocs(manifest Manifest, renderedRoot string) map[string]string {
	docs := map[string]string{manifest.GeneratedDoc: renderedRoot}
	for _, doc := range manifest.DetailDocs {
		docs[doc.Path] = renderDetailDoc(manifest, doc)
	}
	return docs
}
